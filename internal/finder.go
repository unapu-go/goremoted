package internal

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

func Find(timeout time.Duration, dest ...string) (ok string, err error) {
	if timeout == 0 {
		timeout = 4 * time.Second
	}
	var client = http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: timeout,
			}).DialContext,
			TLSHandshakeTimeout: timeout,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	okc, errc := make(chan string), make(chan error)
	var wg sync.WaitGroup
	wg.Add(len(dest))

	var errorS []string
	go func() {
		for err := range errc {
			errorS = append(errorS, err.Error())
		}
	}()

	go func() {
		for okv := range okc {
			if ok == "" {
				ok = okv
			}
		}
	}()

	for _, dest := range dest {
		go func(dest string) {
			defer wg.Done()
			resp, err := client.Head(dest)
			if err != nil {
				errc <- fmt.Errorf("connect to destination %q failed: %s", dest, err)
				return
			}
			switch resp.StatusCode {
			case http.StatusOK:
				okc <- dest
			case http.StatusNotFound:
			default:
				errc <- fmt.Errorf("[destination %s] bad gateway status: %s", dest, resp.Status)
				return
			}
		}(dest)
	}

	wg.Wait()
	close(okc)
	close(errc)

	if ok == "" {
		if len(errorS) > 0 {
			err = errors.New("- " + strings.Join(errorS, "\n- "))
		} else {
			err = os.ErrNotExist
		}
	}
	return
}

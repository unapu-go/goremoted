package internal

import (
	"os"

	"github.com/pkg/errors"

	"github.com/moisespsena-go/logging"
	"github.com/moisespsena-go/signald"

	"github.com/moisespsena-go/httpu"
	"github.com/moisespsena-go/task"
	"github.com/moisespsena-go/xroute"
)

func Run(start chan *Config) (err error) {
	var (
		r   *task.Runner
		cfg *Config
	)

	log := logging.GetOrCreateLogger("main")
	log.Noticef("uses `killall -USR2 %q` to restart it", os.Args[0])

	signald.Restartable()
	signald.Restarts(func(sig os.Signal) {
		start <- cfg
	})

	signald.Done(func(sig os.Signal) {
		close(start)
	})

	signald.Bind(func(signal os.Signal) {
		signald.Stop()
	})

	for cfg = range start {
		if r != nil {
			r.StopWait()
		}

		r = task.NewRunner(&Runner{cfg})
		if _, err = r.Run(); err != nil {
			log.Error(err)
			r = nil
			continue
		}

		signald.AutoBindInterface(r)
	}
	return
}

type Runner struct {
	cfg *Config
}

func (this Runner) Start(done func()) (stop task.Stoper, err error) {
	h := &Handler{this.cfg, map[string]*HostHandler{}}

	func() {
		defer func() {
			if r := recover(); r != nil {
				err = errors.Wrap(r.(error), "register path failed")
			}
		}()

		for host, paths := range this.cfg.Hosts {
			hh := &HostHandler{this.cfg, host, xroute.NewMux()}
			hh.mux.NotFound(h.Fallback)
			for pth := range paths.Patterns {
				hh.mux.Get(pth, hh.ServeHTTPContext)
			}
			h.hosts[host] = hh
		}
	}()

	if err != nil {
		return
	}

	server := httpu.NewServer(&this.cfg.Server, h)
	return task.Start(func(state *task.State) {
		done()
	}, server)
}

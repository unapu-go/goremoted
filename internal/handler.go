package internal

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/moisespsena-go/xroute"
)

type Handler struct {
	cfg   *Config
	hosts map[string]*HostHandler
}

func (this *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host := strings.Split(r.Host, ":")[0]
	if hh, ok := this.hosts[host]; !ok {
		http.NotFound(w, r)
	} else if r.URL.Query().Get("go-get") == "1" {
		hh.mux.ServeHTTP(w, r)
	} else {
		this.Fallback(w, r)
	}
}

func (this *Handler) Fallback(w http.ResponseWriter, r *http.Request) {
	host := strings.Split(r.Host, ":")[0]
	var proto = r.Header.Get("X-Forwarded-Proto")
	if proto == "" {
		proto = "http"
		if r.TLS != nil {
			proto += "https"
		}
	}
	to := strings.ReplaceAll(
		strings.ReplaceAll(
			strings.ReplaceAll(
				strings.ReplaceAll(
					strings.ReplaceAll(
						this.cfg.Fallback.RedirectTo, "MUST_HOST", strings.TrimPrefix(host, "www."),
					), "WWW_HOST", "www."+strings.TrimPrefix(host, "www."),
				),
				"HOST", host,
			),
			"PROTO", proto,
		),
		"URI", strings.TrimPrefix(r.URL.String(), "/"),
	)
	http.Redirect(w, r, to, this.cfg.Fallback.RedirectStatus)
}

type HostHandler struct {
	cfg  *Config
	host string
	mux  *xroute.Mux
}

func (this *HostHandler) ServeHTTPContext(w http.ResponseWriter, r *http.Request, rctx *xroute.RouteContext) {
	parts := strings.SplitN(rctx.URLParam("project"), ".", 2)
	projectName := parts[0]
	var tagName string
	if len(parts) == 2 {
		tagName = parts[1]
	} else {
		tagName = "master"
	}
	pattern := rctx.RoutePattern()
	patternCfg := this.cfg.Hosts[this.host].Patterns[pattern]
	var dests []string
	for _, dest := range patternCfg.Destinations {
		dests = append(dests, fmt.Sprintf(dest, projectName))
	}

	dest, err := Find(this.cfg.HttpClientTimeout, dests...)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "no destination found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusBadGateway)
		}
		return
	}

	w.Header().Set("Content-Type", "text/html")
	pkg := this.host + "/" + projectName
	importContent := fmt.Sprintf("%s git %s", pkg, dest)
	sourceContent := fmt.Sprintf("%s _ %s/tree/%s{/dir} %s/blob/%s{/dir}/{file}#L{line}", pkg, dest, tagName, dest, tagName)
	fmt.Fprintf(w, `<!doctype html>
<html>
	<head>
		<title>%s</title>
		<meta name="go-import" content="%s">
		<meta name="go-source" content="%s">
	</head>
	<body><code>go get -v %s</code></body>
</html>
`, pkg, importContent, sourceContent, pkg)
}

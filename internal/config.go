package internal

import (
	"time"

	"github.com/moisespsena-go/httpu"
)

type Config struct {
	HttpClientTimeout time.Duration `mapstructure:"http_client_timeout"`
	Server            httpu.Config  `mapstructure:"server"`
	Fallback          struct {
		RedirectTo     string `mapstructure:"redirect_to"`
		RedirectStatus int    `mapstructure:"redirect_status"`
	}
	Hosts map[string]struct {
		ProjectPage string `mapstructure:"project_page"`
		Patterns    map[string]struct {
			ProjectPage  string
			Destinations []string `mapstructure:"destinations"`
		} `mapstructure:"patterns"`
	} `mapstructure:"hosts"`
}

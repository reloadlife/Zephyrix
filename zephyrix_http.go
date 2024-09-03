package zephyrix

import (
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

type ServerConfig struct {
	Address     string `mapstructure:"address"`
	TLSEnabled  bool   `mapstructure:"tls_enabled"`
	TLSCertFile string `mapstructure:"tls_cert_file"`
	TLSKeyFile  string `mapstructure:"tls_key_file"`
	TLSAddress  string `mapstructure:"tls_address"`

	TrustedProxies []string `mapstructure:"trusted_proxies"`

	Cors               CorsConfig `mapstructure:"cors"`
	SkipLogPaths       []string   `mapstructure:"skip_log_path"`
	MaxMultipartMemory int64      `mapstructure:"max_multipart_memory"` // todo: make this string and add unit suffixes (parse them)
}

func (z *zephyrix) serveRun(_ *cobra.Command, _ []string) error {
	z.options = append(z.options, fx.Invoke(httpInvoke))
	err := z.fxStart()
	if err != nil {
		return err
	}
	return nil
}

package zephyrix

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

type ServerConfig struct {
	Address string `mapstructure:"address"`

	RedirectToHTTPS bool       `mapstructure:"redirect_to_https"`
	SSL             SSLConfig `mapstructure:"ssl"`

	ReadTimeout  string `mapstructure:"read_timeout"`
	WriteTimeout string `mapstructure:"write_timeout"`
	IdleTimeout  string `mapstructure:"idle_timeout"`

	TrustedProxies []string `mapstructure:"trusted_proxies"`

	Cors               CorsConfig `mapstructure:"cors"`
	SkipLogPaths       []string    `mapstructure:"skip_log_path"`
	MaxMultipartMemory int64       `mapstructure:"max_multipart_memory"` // todo: make this string and add unit suffixes (parse them)
}

// SSLConfig holds all SSL-related configuration options
type SSLConfig struct {
	Enabled              bool     `mapstructure:"enabled"`
	Address              string   `mapstructure:"address"`
	CertFile             string   `mapstructure:"cert_file"`
	KeyFile              string   `mapstructure:"key_file"`
	MinVersion           string   `mapstructure:"min_version"`
	MaxVersion           string   `mapstructure:"max_version"`
	CipherSuites         []string `mapstructure:"cipher_suites"`
	CurvePreferences     []string `mapstructure:"curve_preferences"`
	ClientAuth           string   `mapstructure:"client_auth"`
	ClientCACert         string   `mapstructure:"client_ca_cert"`
	Renegotiation        bool     `mapstructure:"renegotiation"`
	AutoSSLChallangeAddr string   `mapstructure:"auto_ssl_challenge_addr"`
	AutoSSL              bool     `mapstructure:"auto_ssl"`
	AutoSSLDomains       []string `mapstructure:"auto_ssl_domains"`
	AutoSSLEmail         string   `mapstructure:"auto_ssl_email"`
	AutoSSLCacheDir      string   `mapstructure:"auto_ssl_cache_dir"`
	AutoSSLProvider      string   `mapstructure:"auto_ssl_provider"`
	AutoSSLZeroSSLEABKey string   `mapstructure:"auto_ssl_zerossl_eab_key"`
	AutoSSLZeroSSLKID    string   `mapstructure:"auto_ssl_zerossl_kid"`
}

type parsedServerConfig struct {
	*ServerConfig
	ParsedReadTimeout  time.Duration
	ParsedWriteTimeout time.Duration
	ParsedIdleTimeout  time.Duration

	Environment string
}

func (c *ServerConfig) parse(conf *Config) (*parsedServerConfig, error) {
	parsed := &parsedServerConfig{ServerConfig: c}
	parsed.Environment = conf.Environment

	var err error
	parsed.ParsedReadTimeout, err = parseDuration(c.ReadTimeout, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("invalid read_timeout: %w", err)
	}

	parsed.ParsedWriteTimeout, err = parseDuration(c.WriteTimeout, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("invalid write_timeout: %w", err)
	}

	parsed.ParsedIdleTimeout, err = parseDuration(c.IdleTimeout, 120*time.Second)
	if err != nil {
		return nil, fmt.Errorf("invalid idle_timeout: %w", err)
	}

	return parsed, nil
}

func parseDuration(value string, defaultDuration time.Duration) (time.Duration, error) {
	if value == "" {
		return defaultDuration, nil
	}
	return time.ParseDuration(value)
}

func (z *zephyrix) serveRun(_ *cobra.Command, _ []string) error {
	z.options = append(z.options, fx.Invoke(serverInvoke))
	err := z.fxStart()
	if err != nil {
		return err
	}
	return nil
}

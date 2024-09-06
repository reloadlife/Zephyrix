package zephyrix

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Config struct {
	Environment string `mapstructure:"environment"`

	Log      LogConfig      `mapstructure:"log"`
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
}

const (
	configFileName = "zephyrix"
	envPrefix      = "ZEPHYRIX"
)

var (
	configFilePath string
	TestConfig     *Config
)

// initConfig internal function to initialize the configuration and logger based on the configuration
// for testing, when TestConfig is not nil, it will use the TestConfig
func (z *zephyrix) initConfig() {
	if TestConfig != nil {
		z.config = TestConfig
		z.config.Log.Level = "debug" // set log level to debug for tests
		cobra.CheckErr(z.setupLogger())
		return
	}

	v := viper.New()
	z.viper = v

	if configFilePath != "" {
		v.SetConfigFile(configFilePath)
	} else {
		v.SetConfigName(configFileName)
		v.AddConfigPath(".")
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		v.AddConfigPath(home)
	}

	v.SetEnvPrefix(envPrefix)
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			cobra.CheckErr(fmt.Errorf("error reading config file: %w", err))
		}
	}

	cobra.CheckErr(v.Unmarshal(&z.config))
	cobra.CheckErr(z.setupLogger())
}

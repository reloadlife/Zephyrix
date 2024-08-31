package zephyrix

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configFilePath string = "zephyrix.yaml"

type Config struct {
	Log    LogConfig    `mapstructure:"log"`
	Server ServerConfig `mapstructure:"server"`
}

func (z *zephyrix) initConfig() {
	if configFilePath != "" {
		z.viper.SetConfigFile(configFilePath)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		z.viper.AddConfigPath(home)
		z.viper.AddConfigPath(".")
		z.viper.SetConfigName("zephyrix")
	}

	z.viper.SetEnvPrefix("ZEPHYRIX")
	z.viper.AutomaticEnv()
	z.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := z.viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			Logger.Fatal("Error reading config file: %s", err)
		}
	}

	if err := z.viper.Unmarshal(z.config); err != nil {
		Logger.Fatal("Unable to decode config into struct: %s", err)
	}

	cobra.CheckErr(z.setupLogger())
}

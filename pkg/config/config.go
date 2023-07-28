package config

import (
	"fmt"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

func InitConfig(configFilePath string) error {
	viper.SetConfigName("parachute.toml")
	viper.SetConfigType("toml")
	viper.AddConfigPath("/etc/parachute")
	viper.AddConfigPath("$HOME/.config/parachute")
	viper.AddConfigPath(".")

	if configFilePath != "" {
		viper.AddConfigPath(configFilePath)
	}

	viper.AutomaticEnv()
	viper.SetEnvPrefix("parachute")

	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if configFilePath != "" {
				return fmt.Errorf("provided config file '%s' not found: %s", configFilePath, err)
			}
		} else {
			return fmt.Errorf("fatal configuration error: %s", err)
		}
	}

	viper.SetDefault("log_level", zerolog.LevelErrorValue)
	viper.SetDefault("log_format", "")
	viper.SetDefault("passphrase", "")
	viper.SetDefault("no_encryption", false)
	viper.SetDefault("endpoint", "")
	viper.SetDefault("access_key", "")
	viper.SetDefault("secret_key", "")
	viper.SetDefault("output", "")

	return nil
}

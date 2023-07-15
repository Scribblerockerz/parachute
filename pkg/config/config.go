package config

import (
	"fmt"

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

	viper.SetDefault("silent", false)
	viper.SetDefault("passphrase", "")
	viper.SetDefault("no_encryption", false)

	return nil
}

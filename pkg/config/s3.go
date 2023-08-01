package config

import (
	"errors"

	"github.com/spf13/viper"
)

func ValidateS3Configuration() error {
	if viper.GetString("endpoint") == "" {
		return errors.New("endpoint must be provided")
	}

	if viper.GetString("access_key") == "" {
		return errors.New("access key must be provided")
	}

	if viper.GetString("secret_key") == "" {
		return errors.New("secret key must be provided")
	}

	return nil
}

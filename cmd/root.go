package cmd

import (
	"os"

	"github.com/rs/zerolog/log"
	"github.com/scribblerockerz/parachute/pkg/config"
	"github.com/scribblerockerz/parachute/pkg/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configFilePath string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "parachute",
	Short: "A backup utility for s3 compatible storages",
	// 	Long: `A longer description that spans multiple lines and likely contains
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Error().Stack().Err(err).Msg(err.Error())
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(func() {
		err := config.InitConfig(configFilePath)
		if err != nil {
			panic(err)
		}
	})

	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		logger.SetupLogger(viper.GetString("log_level"), viper.GetString("log_format"))
		return nil
	}

	rootCmd.PersistentFlags().StringVarP(&configFilePath, "config", "c", "", "config file (default is parachute.toml)")

	rootCmd.PersistentFlags().String("log-level", "", "verbosity of the output (debug, info, warn, error, fatal, panic)")
	rootCmd.PersistentFlags().String("log-format", "", "logging format (console, json)")
	rootCmd.PersistentFlags().BoolP("no-encryption", "E", false, "prevent archive encryption")
	rootCmd.PersistentFlags().StringP("pass", "p", "", "encryption passphrase")

	viper.BindPFlag("log_level", rootCmd.PersistentFlags().Lookup("log-level"))
	viper.BindPFlag("log_format", rootCmd.PersistentFlags().Lookup("log-format"))
	viper.BindPFlag("no_encryption", rootCmd.PersistentFlags().Lookup("no-encryption"))
	viper.BindPFlag("passphrase", rootCmd.PersistentFlags().Lookup("pass"))
}

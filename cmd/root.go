package cmd

import (
	"os"

	"github.com/scribblerockerz/parachute/pkg/config"
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

	packCmd.PersistentFlags().StringVarP(&configFilePath, "config", "c", "", "config file (default is parachute.toml)")

	packCmd.PersistentFlags().BoolP("silent", "s", false, "prevent human readable output")
	packCmd.PersistentFlags().BoolP("no-encryption", "E", false, "prevent archive encryption")
	packCmd.PersistentFlags().StringP("pass", "p", "", "encryption passphrase")

	viper.BindPFlag("silent", packCmd.PersistentFlags().Lookup("silent"))
	viper.BindPFlag("no_encryption", packCmd.PersistentFlags().Lookup("no-encryption"))
	viper.BindPFlag("passphrase", packCmd.PersistentFlags().Lookup("pass"))
}

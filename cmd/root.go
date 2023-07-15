package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var isSilent bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "parachute",
	Short: "A backup utility for s3 compatible storages",
	// 	Long: `A longer description that spans multiple lines and likely contains
	// examples and usage of using your application. For example:
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	packCmd.PersistentFlags().BoolVarP(&isSilent, "silent", "s", false, "Prevent human readable output")
}

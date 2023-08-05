package version

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Version string
var Revision string
var BuildDate string

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the parachute version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version: %s\nRevision: %s\nBuild Date: %s\n", Version, Revision, BuildDate)
	},
}

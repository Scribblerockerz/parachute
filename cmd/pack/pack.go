package pack

import (
	"errors"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/scribblerockerz/parachute/pkg/archive"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var PackCmd = &cobra.Command{
	Use:    "pack SOURCE [flags]",
	Short:  "Create an archive of a file/directory.",
	RunE:   runPack,
	PreRun: preRun,
}

func init() {
	PackCmd.Flags().StringP("output", "o", "", "output destination")
	PackCmd.Flags().Bool("timed-name", false, "prepend sortable time infront of the archive")
}

// preRun will initialize viper flag bindings, to prevent overrides of the same key
func preRun(cmd *cobra.Command, args []string) {
	viper.BindPFlag("output", cmd.Flags().Lookup("output"))
	viper.BindPFlag("timed_name", cmd.Flags().Lookup("timed-name"))
}

func runPack(cmd *cobra.Command, args []string) error {

	log.Info().Strs("args", args).Msg("started packing")

	err := validatePackInput(args)
	if err != nil {
		return err
	}

	packArgs, err := getPackArgs(args, viper.GetString("output"))
	if err != nil {
		return err
	}

	fmt.Println(packArgs.destination)
	fmt.Println(viper.GetString("output"))

	a, err := archive.CreateArchiveFromSources(
		packArgs.source,
		!viper.GetBool("no_encryption"),
		viper.GetString("passphrase"),
	)
	if err != nil {
		return err
	}

	fileDestination, err := a.CopyIntoDir(a.TempDestination(), packArgs.destination, viper.GetBool("timed_name"))
	if err != nil {
		return err
	}

	err = a.Cleanup()
	if err != nil {
		return err
	}

	log.Info().Str("archive", fileDestination).Msg("finished packing encrypted archive")
	return nil
}

type packArgs struct {
	source      []string
	destination string
}

func getPackArgs(args []string, output string) (*packArgs, error) {
	pathArg := output

	if pathArg == "" {
		cwd, err := os.Getwd()

		if err != nil {
			return nil, err
		}

		pathArg = cwd
	}

	return &packArgs{
		source:      args,
		destination: pathArg,
	}, nil
}

func validatePackInput(args []string) error {
	if len(args) == 0 {
		return errors.New("source file or directory must be provided")
	}

	if !viper.GetBool("no_encryption") && viper.GetString("passphrase") == "" {
		return errors.New("provided passphrase is empty")
	}

	return nil
}

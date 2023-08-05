package unpack

import (
	"errors"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/scribblerockerz/parachute/pkg/archive"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var UnpackCmd = &cobra.Command{
	Use:    "unpack SOURCE [flags]",
	Short:  "Extract an (encrypted) archive to a file/directory.",
	RunE:   runUnpack,
	PreRun: preRun,
}

func init() {
	UnpackCmd.Flags().StringP("output", "o", "", "output destination")
	UnpackCmd.Flags().Bool("timed-name", false, "prepend sortable time infront of the archive")
}

// preRun will initialize viper flag bindings, to prevent overrides of the same key
func preRun(cmd *cobra.Command, args []string) {
	viper.BindPFlag("output", cmd.Flags().Lookup("output"))
}

func runUnpack(cmd *cobra.Command, args []string) error {

	err := validateUnpackInput(args)
	if err != nil {
		panic(err)
	}

	unpackArgs, err := getUnpackArgs(args, viper.GetString("output"))
	if err != nil {
		panic(err)
	}

	a, err := archive.CreateTempArchiveFromRemoteFile(unpackArgs.source)

	if err != nil {
		return err
	}

	_, err = a.CopyIntoDir(unpackArgs.source, a.TempLocation, false)
	if err != nil {
		return err
	}

	if a.IsEncrupted {
		err = a.Decrypt(viper.GetString("passphrase"))
		if err != nil {
			return err
		}

		log.Debug().Str("decryptedFile", a.TempDestination()).Msg("decrypted temporary archive")
	}

	err = a.Unzip()
	if err != nil {
		return err
	}

	fileDestination, err := a.CopyIntoDir(a.Destination(), unpackArgs.destination, false)
	if err != nil {
		return err
	}

	err = a.Cleanup()
	if err != nil {
		return err
	}

	log.Info().Str("destination", fileDestination).Msg("finsihed restore to destination")

	return nil
}

type unpackArgs struct {
	source      string
	destination string
}

func getUnpackArgs(args []string, output string) (*unpackArgs, error) {
	pathArg := output

	if pathArg == "" {
		cwd, err := os.Getwd()

		if err != nil {
			return nil, err
		}

		pathArg = cwd
	}

	return &unpackArgs{
		source:      args[0],
		destination: pathArg,
	}, nil
}

func validateUnpackInput(args []string) error {
	if len(args) == 0 {
		return errors.New("source archive must be provided")
	}

	if strings.HasSuffix(args[0], archive.ENCRYPTED_FILE_SUFFIX) && viper.GetString("passphrase") == "" {
		return errors.New("provided passphrase is empty")
	}

	return nil
}

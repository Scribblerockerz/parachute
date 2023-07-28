package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/scribblerockerz/parachute/pkg/archive"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var unpackCmd = &cobra.Command{
	Use:   "unpack SOURCE [flags]",
	Short: "Extract an (encrypted) archive to a file/directory.",
	Run:   runUnpack,
}

var unpackOutput string

func init() {
	rootCmd.AddCommand(unpackCmd)

	unpackCmd.Flags().StringVarP(&unpackOutput, "output", "o", "", "output destination")
}

func runUnpack(cmd *cobra.Command, args []string) {

	err := validateUnpackInput(args)
	if err != nil {
		panic(err)
	}

	unpackArgs, err := getUnpackArgs(args, unpackOutput)
	if err != nil {
		panic(err)
	}

	isEncrypted := archive.IsFileEncrypted(unpackArgs.source)
	passphrase := viper.GetString("passphrase")

	source := unpackArgs.source
	destination := unpackArgs.destination

	if isEncrypted {

		decryptedDestination, err := archive.DestinationPath(
			destination,
			archive.GetFallbackName(
				source,
				"archive",
				fmt.Sprintf(".%s", archive.ENCRYPTED_FILE_SUFFIX),
			),
			false,
		)
		if err != nil {
			panic(err)
		}

		err = archive.DecryptFile(source, decryptedDestination, passphrase)
		if err != nil {
			panic(err)
		}

		source = decryptedDestination
		destination, _ = strings.CutSuffix(decryptedDestination, ".zip")
	}

	destination, err = archive.DestinationPath(
		destination,
		archive.GetFallbackName(source, "archive", ".zip"),
		false,
	)

	if err != nil {
		panic(err)
	}

	err = archive.UnzipSource(source, destination)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Unpacked data to %s\n", destination)
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

	if strings.HasSuffix(args[0], fmt.Sprintf(".%s", archive.ENCRYPTED_FILE_SUFFIX)) && viper.GetString("passphrase") == "" {
		return errors.New("provided passphrase is empty")
	}

	return nil
}

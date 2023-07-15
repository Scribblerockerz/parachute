package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/scribblerockerz/parachute/pkg/archive"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var packCmd = &cobra.Command{
	Use:   "pack SOURCE [flags]",
	Short: "Create an archive of a file/directory.",
	Run:   run,
}

var output string

func init() {
	rootCmd.AddCommand(packCmd)

	packCmd.Flags().StringVarP(&output, "output", "o", "", "output destination")
}

func run(cmd *cobra.Command, args []string) {

	err := validateInput(args)
	if err != nil {
		panic(err)
	}

	packArgs, err := getPackArgs(args, output)
	if err != nil {
		panic(err)
	}

	archivePath, err := archive.NewArchive(packArgs.source, packArgs.destination)
	if err != nil {
		panic(err)
	}

	if viper.GetBool("no_encryption") {
		fmt.Printf("Packed UNENCRYPTED data to %s\n", archivePath)
		return
	}

	encryptedFile := fmt.Sprintf("%s.enc", archivePath)
	passphrase := viper.GetString("passphrase")

	err = archive.EncryptFile(archivePath, encryptedFile, passphrase)
	if err != nil {
		panic(err)
	}

	err = os.Remove(archivePath)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Packed encrypted data to %s\n", encryptedFile)
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

func validateInput(args []string) error {
	if len(args) == 0 {
		return errors.New("source file or directory must be provided")
	}

	if !viper.GetBool("no_encryption") && viper.GetString("passphrase") == "" {
		return errors.New("provided passphrase is empty")
	}

	return nil
}
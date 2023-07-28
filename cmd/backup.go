package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/scribblerockerz/parachute/pkg/archive"
	"github.com/scribblerockerz/parachute/pkg/s3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var backupCmd = &cobra.Command{
	Use:   "backup SOURCE [flags]",
	Short: "Create an archive (encrypted) of the SOURCE and move it to the destination",
	RunE:  runBackup,
}

func init() {
	rootCmd.AddCommand(backupCmd)

	backupCmd.Flags().StringP("output", "o", "", "output destination")
	backupCmd.Flags().String("endpoint", "", "s3 endpoint")
	backupCmd.Flags().String("access-key", "", "s3 access key")
	backupCmd.Flags().String("secret-key", "", "s3 secret key")

	viper.BindPFlag("endpoint", backupCmd.Flags().Lookup("endpoint"))
	viper.BindPFlag("access_key", backupCmd.Flags().Lookup("access-key"))
	viper.BindPFlag("secret_key", backupCmd.Flags().Lookup("secret-key"))
	viper.BindPFlag("output", backupCmd.Flags().Lookup("output"))
}

func runBackup(cmd *cobra.Command, args []string) error {

	log.Info().Strs("args", args).Msg("starting backup crreation")

	err := validateBackupInput(args, viper.GetString("output"))
	if err != nil {
		return err
	}

	backupArgs, err := getBackupArgs(args, viper.GetString("output"))
	if err != nil {
		return err
	}

	err = validateS3Configuration()
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	archivePath, err := archive.DestinationPath(
		cwd,
		"archive.zip",
		false,
	)
	if err != nil {
		return err
	}

	err = archive.ZipSource(backupArgs.source, archivePath)
	if err != nil {
		return err
	}

	log.Debug().Str("archive", archivePath).Msg("created temporary archive")

	uploadSourceFile := archivePath

	if !viper.GetBool("no_encryption") {
		encryptedFile := fmt.Sprintf("%s.%s", archivePath, archive.ENCRYPTED_FILE_SUFFIX)
		passphrase := viper.GetString("passphrase")

		err = archive.EncryptFile(archivePath, encryptedFile, passphrase)
		if err != nil {
			return err
		}

		log.Debug().Str("encryptedFile", encryptedFile).Msg("encrypted temporary archive")

		err = os.Remove(archivePath)
		if err != nil {
			return err
		}

		log.Debug().Str("archive", archivePath).Msg("removed temporary unencrypted archive")

		uploadSourceFile = encryptedFile
	}

	client, err := s3.NewClient(
		viper.GetString("endpoint"),
		viper.GetString("access_key"),
		viper.GetString("secret_key"),
		true,
	)

	if err != nil {
		return err
	}

	payload, err := s3.NewPayload(backupArgs.destination, uploadSourceFile)
	if err != nil {
		return err
	}

	log.Debug().Str("bucket", payload.Bucket).Str("object", payload.Object).Msg("started uploading")

	_, err = client.UploadPayload(
		context.Background(),
		payload,
	)

	if err != nil {
		return err
	}

	log.Debug().Str("bucket", payload.Bucket).Str("object", payload.Object).Msg("finished uploading")

	err = os.Remove(uploadSourceFile)
	if err != nil {
		return err
	}

	log.Debug().Str("archive", uploadSourceFile).Msg("removed temporary archive")
	log.Info().Str("destination", backupArgs.destination).Msg("finsihed backup to destination")

	return nil
}

func validateBackupInput(args []string, output string) error {
	if len(args) == 0 {
		return errors.New("source archive must be provided")
	}

	if !viper.GetBool("no_encryption") && viper.GetString("passphrase") == "" {
		return errors.New("provided passphrase is empty")
	}

	if viper.GetBool("no_encryption") {
		log.Warn().Msg("no encryption requested")
	}

	if output == "" {
		return errors.New("output destination must be provided")
	}

	if !strings.HasPrefix(output, "s3://") {
		return errors.New("output must be declared in \"s3://bucket/some-path\" format")
	}

	return nil
}

type backupArgs struct {
	source      []string
	destination string
}

func getBackupArgs(args []string, output string) (*backupArgs, error) {
	return &backupArgs{
		source:      args,
		destination: output,
	}, nil
}

func validateS3Configuration() error {
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

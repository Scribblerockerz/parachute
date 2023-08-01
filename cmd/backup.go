package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/scribblerockerz/parachute/pkg/archive"
	"github.com/scribblerockerz/parachute/pkg/config"
	"github.com/scribblerockerz/parachute/pkg/s3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var backupCmd = &cobra.Command{
	Use:   "backup LOCAL [flags]",
	Short: "Create an archive (encrypted) of the LOCAL souce and move it to the REMOTE destination",
	RunE:  runBackup,
}

func init() {
	rootCmd.AddCommand(backupCmd)

	backupCmd.Flags().StringP("remote", "r", "", "remote destination (S3)")
	backupCmd.Flags().String("endpoint", "", "S3 endpoint")
	backupCmd.Flags().String("access-key", "", "S3 access key")
	backupCmd.Flags().String("secret-key", "", "S3 secret key")

	viper.BindPFlag("endpoint", backupCmd.Flags().Lookup("endpoint"))
	viper.BindPFlag("access_key", backupCmd.Flags().Lookup("access-key"))
	viper.BindPFlag("secret_key", backupCmd.Flags().Lookup("secret-key"))
	viper.BindPFlag("remote", backupCmd.Flags().Lookup("remote"))
}

func runBackup(cmd *cobra.Command, args []string) error {

	log.Info().Strs("args", args).Msg("started backup creation")

	err := validateBackupInput(args, viper.GetString("remote"))
	if err != nil {
		return err
	}

	backupArgs, err := getBackupArgs(args, viper.GetString("remote"))
	if err != nil {
		return err
	}

	err = config.ValidateS3Configuration()
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
		encryptedFile := fmt.Sprintf("%s%s", archivePath, archive.ENCRYPTED_FILE_SUFFIX)
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

func validateBackupInput(args []string, remote string) error {
	if len(args) == 0 {
		return errors.New("source archive must be provided")
	}

	if !viper.GetBool("no_encryption") && viper.GetString("passphrase") == "" {
		return errors.New("provided passphrase is empty")
	}

	if viper.GetBool("no_encryption") {
		log.Warn().Msg("no encryption requested")
	}

	if remote == "" {
		return errors.New("remote destination must be provided")
	}

	if !strings.HasPrefix(remote, "s3://") {
		return errors.New("remote must be declared in \"s3://bucket/some-path\" format")
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

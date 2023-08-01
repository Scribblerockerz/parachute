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

var restoreCmd = &cobra.Command{
	Use:   "restore LOCAL [flags]",
	Short: "Restore an REMOTE archive (encrypted) from an S3 source and move it to a LOCAL destination",
	RunE:  runRestore,
}

func init() {
	rootCmd.AddCommand(restoreCmd)

	restoreCmd.Flags().StringP("remote", "o", "", "remote destination (S3)")
	restoreCmd.Flags().String("endpoint", "", "S3 endpoint")
	restoreCmd.Flags().String("access-key", "", "S3 access key")
	restoreCmd.Flags().String("secret-key", "", "S3 secret key")

	viper.BindPFlag("endpoint", restoreCmd.Flags().Lookup("endpoint"))
	viper.BindPFlag("access_key", restoreCmd.Flags().Lookup("access-key"))
	viper.BindPFlag("secret_key", restoreCmd.Flags().Lookup("secret-key"))
	viper.BindPFlag("remote", restoreCmd.Flags().Lookup("remote"))
}

func runRestore(cmd *cobra.Command, args []string) error {

	log.Info().Strs("args", args).Msg("started restoring")

	err := validateRestoreInput(args, viper.GetString("remote"))
	if err != nil {
		return err
	}

	restoreArgs, err := getRestoreArgs(args, viper.GetString("remote"))
	if err != nil {
		return err
	}

	err = config.ValidateS3Configuration()
	if err != nil {
		return err
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

	destinationFilePath := fmt.Sprintf("%s.zip", restoreArgs.destination)
	isEncrypted := archive.IsFileEncrypted(restoreArgs.remote)

	if isEncrypted {
		destinationFilePath = fmt.Sprintf("%s%s", destinationFilePath, archive.ENCRYPTED_FILE_SUFFIX)
	}

	downloadInfo, err := s3.NewDownload(restoreArgs.remote, destinationFilePath)
	if err != nil {
		return err
	}

	log.Debug().Str("bucket", downloadInfo.Bucket).Str("object", downloadInfo.Object).Str("filePath", downloadInfo.FilePath).Msg("started downloading")

	err = client.DownloadPayload(
		context.Background(),
		downloadInfo,
	)

	if err != nil {
		return err
	}

	log.Debug().Str("bucket", downloadInfo.Bucket).Str("object", downloadInfo.Object).Msg("finished downloading")

	archivePath := downloadInfo.FilePath

	if isEncrypted {
		passphrase := viper.GetString("passphrase")
		decryptedFile, _ := strings.CutSuffix(downloadInfo.FilePath, archive.ENCRYPTED_FILE_SUFFIX)
		err := archive.DecryptFile(downloadInfo.FilePath, decryptedFile, passphrase)
		if err != nil {
			return err
		}

		log.Debug().Str("decryptedFile", decryptedFile).Msg("decrypted temporary archive")

		archivePath = decryptedFile

		err = os.Remove(downloadInfo.FilePath)
		if err != nil {
			return err
		}

		log.Debug().Str("encryptedFile", downloadInfo.FilePath).Msg("removed temporary encrypted archive")
	}

	err = archive.UnzipSource(archivePath, restoreArgs.destination)
	if err != nil {
		return err
	}

	err = os.Remove(archivePath)
	if err != nil {
		return err
	}

	log.Debug().Str("archive", archivePath).Msg("removed temporary archive")
	log.Info().Str("destination", restoreArgs.destination).Msg("finsihed restore to destination")

	return nil
}

func validateRestoreInput(args []string, remote string) error {
	if len(args) != 1 {
		return errors.New("archive destination must be provided")
	}

	if !viper.GetBool("no_encryption") && viper.GetString("passphrase") == "" {
		return errors.New("provided passphrase is empty")
	}

	if viper.GetBool("no_encryption") {
		log.Warn().Msg("no encryption requested")
	}

	if remote == "" {
		return errors.New("remote source must be provided")
	}

	if !strings.HasPrefix(remote, "s3://") {
		return errors.New("remote must be declared in \"s3://bucket/some-path\" format")
	}

	isEncrypted := archive.IsFileEncrypted(remote)

	if isEncrypted && viper.GetString("passphrase") == "" {
		return fmt.Errorf("remote object contains encryption hint (%s) but passphrase is empty", archive.ENCRYPTED_FILE_SUFFIX)
	}

	if isEncrypted && viper.GetBool("no_encryption") {
		log.Warn().Str("remote", remote).Msg(fmt.Sprintf("remote object contains encryption hint (%s) but configured to 'not encrypt'", archive.ENCRYPTED_FILE_SUFFIX))
	}

	return nil
}

type restoreArgs struct {
	destination string
	remote      string
}

func getRestoreArgs(args []string, remote string) (*restoreArgs, error) {
	return &restoreArgs{
		destination: args[0],
		remote:      remote,
	}, nil
}

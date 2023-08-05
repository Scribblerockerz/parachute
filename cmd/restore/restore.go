package restore

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/scribblerockerz/parachute/pkg/archive"
	"github.com/scribblerockerz/parachute/pkg/config"
	"github.com/scribblerockerz/parachute/pkg/s3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var RestoreCmd = &cobra.Command{
	Use:    "restore LOCAL [flags]",
	Short:  "Restore an REMOTE archive (encrypted) from an S3 source and move it to a LOCAL destination",
	RunE:   runRestore,
	PreRun: preRun,
}

func init() {
	RestoreCmd.Flags().StringP("remote", "o", "", "remote destination (S3)")
	RestoreCmd.Flags().String("endpoint", "", "S3 endpoint")
	RestoreCmd.Flags().String("access-key", "", "S3 access key")
	RestoreCmd.Flags().String("secret-key", "", "S3 secret key")
}

// preRun will initialize viper flag bindings, to prevent overrides of the same key
func preRun(cmd *cobra.Command, args []string) {
	viper.BindPFlag("endpoint", cmd.Flags().Lookup("endpoint"))
	viper.BindPFlag("access_key", cmd.Flags().Lookup("access-key"))
	viper.BindPFlag("secret_key", cmd.Flags().Lookup("secret-key"))
	viper.BindPFlag("remote", cmd.Flags().Lookup("remote"))
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

	a, err := archive.CreateTempArchiveFromRemoteFile(restoreArgs.remote)

	if err != nil {
		return err
	}

	downloadInfo, err := s3.NewDownload(restoreArgs.remote, a.TempDestination())
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

	log.Debug().Str("bucket", downloadInfo.Bucket).Str("object", downloadInfo.Object).Str("temdDestination", a.TempDestination()).Msg("finished downloading")

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

	fileDestination, err := a.CopyIntoDir(a.Destination(), restoreArgs.destination, false)
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

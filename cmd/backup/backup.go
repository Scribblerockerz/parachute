package backup

import (
	"context"
	"errors"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/scribblerockerz/parachute/pkg/archive"
	"github.com/scribblerockerz/parachute/pkg/config"
	"github.com/scribblerockerz/parachute/pkg/s3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var BackupCmd = &cobra.Command{
	Use:    "backup LOCAL [flags]",
	Short:  "Create an archive (encrypted) of the LOCAL souce and move it to the REMOTE destination",
	RunE:   runBackup,
	PreRun: preRun,
}

func init() {
	BackupCmd.Flags().StringP("remote", "r", "", "remote destination (S3)")
	BackupCmd.Flags().String("endpoint", "", "S3 endpoint")
	BackupCmd.Flags().String("access-key", "", "S3 access key")
	BackupCmd.Flags().String("secret-key", "", "S3 secret key")
}

// preRun will initialize viper flag bindings, to prevent overrides of the same key
func preRun(cmd *cobra.Command, args []string) {
	viper.BindPFlag("endpoint", cmd.Flags().Lookup("endpoint"))
	viper.BindPFlag("access_key", cmd.Flags().Lookup("access-key"))
	viper.BindPFlag("secret_key", cmd.Flags().Lookup("secret-key"))
	viper.BindPFlag("remote", cmd.Flags().Lookup("remote"))
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

	a, err := archive.CreateArchiveFromSources(
		backupArgs.source,
		!viper.GetBool("no_encryption"),
		viper.GetString("passphrase"),
	)
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

	payload, err := s3.NewPayload(backupArgs.destination, a.TempDestination())
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

	err = a.Cleanup()
	if err != nil {
		return err
	}

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

package logger

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func SetupLogger(level string, format string) error {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	if format != "json" {
		log.Logger = zerolog.New(getOpinionatedConsoleWriter()).With().Timestamp().Logger()
	}

	if level == "" {
		return nil
	}

	zerologLevel, err := zerolog.ParseLevel(level)

	if err != nil {
		return err
	}

	zerolog.SetGlobalLevel(zerologLevel)

	return nil
}

func getOpinionatedConsoleWriter() zerolog.ConsoleWriter {
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006-01-02 15:04:05"}
	output.NoColor = true
	output.FormatLevel = func(i interface{}) string {
		return strings.ToUpper(fmt.Sprintf("\t%-6s\t", i))
	}
	output.FormatMessage = func(i interface{}) string {
		return fmt.Sprintf("%s\t", i)
	}
	return output
}

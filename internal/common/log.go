package common

import (
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
	"io"
	"os"
)

// GetLogFlags return the logging flags
func GetLogFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{Name: "log-level", Usage: "the logging level", Value: "info"},
		&cli.StringFlag{Name: "log-file", Usage: "path to the log file"},
	}
}

// ConfigureLogger configure zerolog.Logger using given cli.Context
func ConfigureLogger(c *cli.Context) (zerolog.Logger, error) {
	// Parse log level
	lvl, err := zerolog.ParseLevel(c.String("log-level"))
	if err != nil {
		return zerolog.Logger{}, err
	}

	var writers []io.Writer
	writer := zerolog.NewConsoleWriter()
	writers = append(writers, writer)

	if file := c.String("log-file"); file != "" {
		f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0640)
		if err != nil {
			return zerolog.Logger{}, err
		}
		writers = append(writers, f)
	}

	l := zerolog.New(zerolog.MultiLevelWriter(writers...)).
		With().
		Timestamp().
		Logger()

	return l.Level(lvl), nil
}

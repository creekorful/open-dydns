package common

import (
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
	"os"
)

// GetLogFlag return a pointer to a cli.StringFlag
// to use for configuring logging
func GetLogFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:  "log-level",
		Value: "info",
	}
}

// ConfigureLogger configure zerolog.Logger using given cli.Context
func ConfigureLogger(c *cli.Context) (zerolog.Logger, error) {
	// Parse log level
	lvl, err := zerolog.ParseLevel(c.String("log-level"))
	if err != nil {
		return zerolog.Logger{}, err
	}

	writer := zerolog.NewConsoleWriter()
	writer.Out = os.Stdout

	l := zerolog.New(writer).
		With().
		Timestamp().
		Logger()

	return l.Level(lvl), nil
}

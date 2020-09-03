package common

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"os"
)

func GetLogFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:  "log-level",
		Value: "info",
	}
}

func ConfigureLogger(c *cli.Context) error {
	lvl, err := zerolog.ParseLevel(c.String("log-level"))
	if err != nil {
		return err
	}

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout}).Level(lvl)
	return nil
}

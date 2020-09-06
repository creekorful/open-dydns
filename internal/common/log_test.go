package common

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
	"testing"
)

func TestGetLogFlags(t *testing.T) {
	flags := GetLogFlags()

	if len(flags) != 2 {
		t.Error("Wrong number of flags returned")
	}

	// TODO assert on specific flag
}

func TestConfigureLogger(t *testing.T) {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "log-level", Value: "info"},
		},
		Action: run,
	}

	if err := app.Run([]string{"app", "--log-level=debug"}); err != nil {
		t.Errorf("ConfigureLogger() has failed: %s", err)
	}
}

func run(c *cli.Context) error {
	l, err := ConfigureLogger(c)
	if err != nil {
		return err
	}

	if l.GetLevel() != zerolog.DebugLevel {
		return fmt.Errorf("wrong log level: %s", l.GetLevel().String())
	}

	return nil
}

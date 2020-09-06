package main

import (
	"github.com/creekorful/open-dydns/internal/opendydnscli"
	"os"
)

func main() {
	if err := opendydnscli.NewCLIApp().App().Run(os.Args); err != nil {
		os.Exit(1)
	}
}

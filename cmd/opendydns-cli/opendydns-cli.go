package main

import (
	"github.com/creekorful/open-dydns/internal/opendydnscli"
	"os"
)

func main() {
	if err := opendydnscli.NewCLI().App().Run(os.Args); err != nil {
		os.Exit(1)
	}
}

package main

import (
	"github.com/creekorful/open-dydns/internal/opendydnsctl"
	"os"
)

func main() {
	if err := opendydnsctl.NewCLIApp().App().Run(os.Args); err != nil {
		os.Exit(1)
	}
}

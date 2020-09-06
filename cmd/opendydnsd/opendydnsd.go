package main

import (
	"github.com/creekorful/open-dydns/internal/opendydnsd"
	"os"
)

func main() {
	if err := opendydnsd.NewDaemonApp().GetApp().Run(os.Args); err != nil {
		os.Exit(1)
	}
}

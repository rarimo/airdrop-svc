package main

import (
	"os"

	"github.com/rarimo/airdrop-svc/internal/cli"
)

func main() {
	if !cli.Run(os.Args) {
		os.Exit(1)
	}
}

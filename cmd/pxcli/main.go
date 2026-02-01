package main

import (
	"os"

	"pxcli/internal/buildinfo"
	"pxcli/internal/cli"
)

func main() {
	cmd := cli.NewRootCmd(buildinfo.Version)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

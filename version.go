package main

import (
	"fmt"
	"runtime"

	"github.com/cloud66/cli"
)

var cmdVersion = &Command{
	Name:       "version",
	Build:      buildBasicCommand,
	Run:        runVersion,
	Short:      "shows cx version",
	NeedsStack: false,
	NeedsOrg:   false,
}

func runVersion(c *cli.Context) {
	fmt.Println(VERSION)
	if debugMode {
		fmt.Println("Running in debug mode")
		if BUILD_DATE == "" {
			fmt.Println("Build date: unknown")
		} else {
			fmt.Printf("Build date: %s\n", BUILD_DATE)
		}

		fmt.Printf("Go Version: %s\n", runtime.Version())
	}
}

package main

import (
	"fmt"

	"github.com/cloud66/cli"
)

var cmdTest = &Command{
	Name:       "test",
	Build:      buildBasicCommand,
	Run:        runTest,
	NeedsStack: false,
	Short:      "checks if cx is properly installed and authenticated",
	Long: `This command is usually used by other tools to check if Cloud 66 Toolbelt is installed, configured and authenticated with the server.
`,
}

func runTest(c *cli.Context) {
	err := client.AuthenticatedPing()
	if err != nil {
		printFatal("Authentication failed: %s", err.Error())
	}

	fmt.Println("Authenticated")
}

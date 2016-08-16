package main

import (
	"os"

	"github.com/cloud66/cli"
)

func runServicePause(c *cli.Context) {
	if len(c.Args()) != 1 {
		cli.ShowSubcommandHelp(c)
		os.Exit(2)
	}

	stack := mustStack(c)
	serviceName := c.Args()[0]
	serverUid := findServerUid(*stack, c)

	asyncId, err := startServiceAction(stack.Uid, &serviceName, serverUid, "service_pause")
	if err != nil {
		printFatal(err.Error())
	}
	genericRes, err := endServiceAction(*asyncId, stack.Uid)
	if err != nil {
		printFatal(err.Error())
	}
	printGenericResponse(*genericRes)
	return
}

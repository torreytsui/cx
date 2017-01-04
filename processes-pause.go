package main

import (
	"os"

	"github.com/cloud66/cli"
)

func runProcessPause(c *cli.Context) {
	if len(c.Args()) > 1 {
		cli.ShowSubcommandHelp(c)
		os.Exit(2)
	}

	// get stack
	stack := mustStack(c)

	// get serverUID
	var serverUID *string
	flagServer := c.String("server")
	if flagServer != "" {
		server := mustServer(c, *stack, flagServer, true)
		serverUID = &server.Uid
		} else {
			if len(c.Args()) == 0 {
				cli.ShowSubcommandHelp(c)
				os.Exit(2)
			}
		}

	// get processName
	var processName *string
	if len(c.Args()) != 0 {
		flagProcess := c.Args()[0]
		if flagProcess != "" {
			processName = &flagProcess
		}
	}

	asyncId, err := startProcessAction(stack.Uid, processName, serverUID, "process_pause")
	if err != nil {
		printFatal(err.Error())
	}
	genericRes, err := endProcessAction(*asyncId, stack.Uid)
	if err != nil {
		printFatal(err.Error())
	}
	printGenericResponse(*genericRes)
	return
}

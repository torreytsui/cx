package main

import (
	"github.com/cloud66/cli"
)

func runProcessRestart(c *cli.Context) {
	// get stack
	stack := mustStack(c)

	// get serverUID
	var serverUID *string
	flagServer := c.String("server")
	if flagServer != "" {
		server := mustServer(c, *stack, flagServer, true)
		serverUID = &server.Uid
	}

	// get processName
	var processName *string
	flagProcess := c.String("process")
	if flagProcess != "" {
		processName = &flagProcess
	}

	asyncId, err := startProcessAction(stack.Uid, processName, serverUID, "process_restart")
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

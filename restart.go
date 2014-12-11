package main

import (
	"os"
	"time"

	"github.com/cloud66/cloud66"
)

var cmdRestart = &Command{
	Run:        runRestart,
	Usage:      "restart",
	NeedsStack: true,
	Category:   "stack",
	Short:      "restarts the stack.",
	Long: `This will send a restart method to all stack components. This means different things for different components.
For a web server, it means a restart of nginx. For an application server, this might be a restart of the workers like Unicorn.
For more information on restart command, please refer to help.cloud66.com
`,
}

func runRestart(cmd *Command, args []string) {
	if len(args) != 0 {
		cmd.printUsage()
		os.Exit(2)
	}
	stack := mustStack()
	asyncId, err := startRestart(stack.Uid)
	if err != nil {
		printFatal(err.Error())
	}
	genericRes, err := endRestart(*asyncId, stack.Uid)
	if err != nil {
		printFatal(err.Error())
	}
	printGenericResponse(*genericRes)
}

func startRestart(stackUid string) (*int, error) {
	asyncRes, err := client.InvokeStackAction(stackUid, "restart")
	if err != nil {
		return nil, err
	}
	return &asyncRes.Id, err
}

func endRestart(asyncId int, stackUid string) (*cloud66.GenericResponse, error) {
	return client.WaitStackAsyncAction(asyncId, stackUid, 5*time.Second, 20*time.Minute, true)
}

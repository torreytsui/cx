package main

import (
	"time"

	"github.com/cloud66/cloud66"

	"github.com/codegangsta/cli"
)

func runRestart(c *cli.Context) {
	stack := mustStack(c)
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

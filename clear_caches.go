package main

import (
	"time"

	"github.com/cloud66/cloud66"

	"github.com/codegangsta/cli"
)

func runClearCaches(c *cli.Context) {
	stack := mustStack(c)
	asyncId, err := startClearCaches(stack.Uid)
	if err != nil {
		printFatal(err.Error())
	}
	genericRes, err := endClearCaches(*asyncId, stack.Uid)
	if err != nil {
		printFatal(err.Error())
	}
	printGenericResponse(*genericRes)
}

func startClearCaches(stackUid string) (*int, error) {
	asyncRes, err := client.InvokeStackAction(stackUid, "clear_caches")
	if err != nil {
		return nil, err
	}
	return &asyncRes.Id, err
}

func endClearCaches(asyncId int, stackUid string) (*cloud66.GenericResponse, error) {
	return client.WaitStackAsyncAction(asyncId, stackUid, 5*time.Second, 20*time.Minute, false)
}

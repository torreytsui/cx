package main

import (
	"fmt"
	"os"
	"time"

	"github.com/cloud66/cloud66"

	"github.com/cloud66/cli"
)

func runJobRun(c *cli.Context) {

	stack := mustStack(c)

	// get the job
	serverName := c.String("job")
	if len(jobName) == 0 {
		cli.ShowSubcommandHelp(c)
		os.Exit(2)
	}

	// asyncId, err := startJobRun(stack.Uid, serviceName, serverUid)
	// if err != nil {
	// 	printFatal(err.Error())
	// }
	// genericRes, err := endJobRun(*asyncId, stack.Uid)
	// if err != nil {
	// 	printFatal(err.Error())
	// }
	// printGenericResponse(*genericRes)
	return
}

func startJobRun(stackUid string, serviceName string, serverUid *string) (*int, error) {
	asyncRes, err := client.StopService(stackUid, serviceName, serverUid)
	if err != nil {
		return nil, err
	}
	return &asyncRes.Id, err
}

func endJobRun(asyncId int, stackUid string) (*cloud66.GenericResponse, error) {
	return client.WaitStackAsyncAction(asyncId, stackUid, 5*time.Second, 20*time.Minute, true)
}

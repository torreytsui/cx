package main

import (
	"fmt"
	"os"
	"time"

	"github.com/cloud66-oss/cloud66"

	"github.com/cloud66/cli"
)

func runSet(c *cli.Context) {
	if len(c.Args()) != 2 {
		//cmd.printUsage()
		os.Exit(2)
	}

	key := c.Args()[0]
	value := c.Args()[1]

	stack := mustStack(c)
	settings, err := client.StackSettings(stack.Uid)
	must(err)

	// check to see if it's a valid setting
	for _, i := range settings {
		if key == i.Key {
			// yup. it's a good one
			fmt.Printf("Please wait while your setting is applied...\n")

			asyncId, err := startSet(stack.Uid, key, value)
			if err != nil {
				printFatal(err.Error())
			}
			genericRes, err := endSet(*asyncId, stack.Uid)
			if err != nil {
				printFatal(err.Error())
			}
			printGenericResponse(*genericRes)

			return
		}
	}

	printFatal(key + " is not a valid setting or does not apply to this stack")
}

func startSet(stackUid string, key string, value string) (*int, error) {
	asyncRes, err := client.Set(stackUid, key, value)
	if err != nil {
		return nil, err
	}
	return &asyncRes.Id, err
}

func endSet(asyncId int, stackUid string) (*cloud66.GenericResponse, error) {
	return client.WaitStackAsyncAction(asyncId, stackUid, 5*time.Second, 20*time.Minute, true)
}

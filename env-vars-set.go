package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cloud66-oss/cloud66"

	"github.com/cloud66/cli"
)

func runEnvVarsSet(c *cli.Context) {
	if len(c.Args()) != 1 {
		cli.ShowSubcommandHelp(c)
		os.Exit(2)
	}

	kv := c.Args()[0]
	kvs := strings.Split(kv, "=")

	if len(kvs) < 2 {
		cli.ShowSubcommandHelp(c)
		os.Exit(2)
	}

	key := kvs[0]
	value := strings.Join(kvs[1:], "=")

	stack := mustStack(c)

	envVars, err := client.StackEnvVars(stack.Uid)
	must(err)

	existing := false
	for _, i := range envVars {
		if i.Key == key {
			if i.Readonly == true {
				printFatal("The selected environment variable is readonly")
			} else {
				existing = true
			}
		}
	}

	fmt.Println("Please wait while your environment variable setting is applied...")

	asyncId, err := startEnvVarSet(stack.Uid, key, value, existing)
	if err != nil {
		printFatal(err.Error())
	}
	genericRes, err := endEnvVarSet(*asyncId, stack.Uid)
	if err != nil {
		printFatal(err.Error())
	}
	printGenericResponse(*genericRes)

	return
}

func startEnvVarSet(stackUid string, key string, value string, existing bool) (*int, error) {
	var (
		asyncRes *cloud66.AsyncResult
		err      error
	)
	if existing {
		asyncRes, err = client.StackEnvVarSet(stackUid, key, value)
	} else {
		asyncRes, err = client.StackEnvVarNew(stackUid, key, value)
	}
	if err != nil {
		return nil, err
	}
	return &asyncRes.Id, err
}

func endEnvVarSet(asyncId int, stackUid string) (*cloud66.GenericResponse, error) {
	return client.WaitStackAsyncAction(asyncId, stackUid, 3*time.Second, 20*time.Minute, true)
}

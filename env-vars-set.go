// +build ignore

package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cloud66/cloud66"
)

var cmdEnvVarsSet = &Command{
	Run:        runEnvVarsSet,
	Usage:      "env-vars-set <setting> <value>",
	NeedsStack: true,
	Category:   "stack",
	Short:      "sets the value of an environment variable on a stack",
	Long: `This sets and applies the value of an environment variable on a stack.
This work happens in the background, therefore this command will return immediately after the operation has started.
Warning! Applying environment variable changes to your stack will result in all your stack environment variables
being sent to your stack servers, and your processes being restarted immediately.

Examples:
$ cx env-var-set -s mystack FIRST_VAR 123
$ cx env-var-set -s mystack SECOND_ONE 'this value has a space in it'
`,
}

func runEnvVarsSet(cmd *Command, args []string) {
	if len(args) != 2 {
		cmd.printUsage()
		os.Exit(2)
	}

	key := strings.ToUpper(args[0])
	value := args[1]

	stack := mustStack()

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

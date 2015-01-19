package main

import (
	"fmt"
	"os"
	"time"

	"github.com/cloud66/cloud66"
)

var cmdSet = &Command{
	Run:        runSet,
	Usage:      "set <setting> <value>",
	NeedsStack: true,
	Category:   "stack",
	Short:      "sets the value of a setting on a stack",
	Long: `This sets and applies the value of a setting on a stack. Applying some settings might require more
work and therefore this command will return immediately after the setting operation has started.

Examples:
$ cx set -s mystack git.branch dev
$ cx set -s mystack allowed.web.source 191.203.12.10
$ cx set -s mystack allowed.web.source anywhere
$ cx set -s mystack maintenance.mode  1|true|on|enable
$ cx set -s mystack maintenance.mode  0|false|off|disable
`,
}

func runSet(cmd *Command, args []string) {
	if len(args) != 2 {
		cmd.printUsage()
		os.Exit(2)
	}

	key := args[0]
	value := args[1]

	stack := mustStack()
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

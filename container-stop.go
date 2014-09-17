package main

import (
	"os"
	"text/tabwriter"
	"time"

	"github.com/cloud66/cloud66"
)

var cmdContainerStop = &Command{
	Run:        runContainerStop,
	Usage:      "container-stop <container id>",
	NeedsStack: true,
	Category:   "stack",
	Short:      "Stops a particular container on the given stack",
	Long: `Stops a particular container on the given stack by container Id.

Examples:
$ cx container-stop -s mystack 2844142cbfc064123777b6be765b3914e43a9e083afce4e4348b5979127c220c
`,
}

func runContainerStop(cmd *Command, args []string) {
	if len(args) == 0 {
		cmd.printUsage()
		os.Exit(2)
	}

	stack := mustStack()
	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()

	containerUid := args[0]
	container, err := client.GetContainer(stack.Uid, containerUid)
	must(err)

	if container == nil {
		printFatal("Container with Id '" + containerUid + "' not found")
	}

	asyncId, err := startContainerStop(stack.Uid, containerUid)
	if err != nil {
		printFatal(err.Error())
	}
	genericRes, err := endServerSet(*asyncId, stack.Uid)
	if err != nil {
		printFatal(err.Error())
	}
	printGenericResponse(*genericRes)
	return
}

func startContainerStop(stackUid string, containerUid string) (*int, error) {
	asyncRes, err := client.StopContainer(stackUid, containerUid)
	if err != nil {
		return nil, err
	}
	return &asyncRes.Id, err
}

func endContainerStop(asyncId int, stackUid string) (*cloud66.GenericResponse, error) {
	return client.WaitStackAsyncAction(asyncId, stackUid, 3*time.Second, 20*time.Minute)
}

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/cloud66-oss/cloud66"

	"github.com/cloud66/cli"
)

func runStackReboot(c *cli.Context) {
	stack := mustStack(c)
	if len(c.Args()) != 0 {
		cli.ShowSubcommandHelp(c)
		os.Exit(2)
	}

	// confirmation is needed if the stack is production
	if !c.Bool("y") {
		mustConfirm("This operation will reboot one or more servers from your stack; during this time your server may not be available. Proceed with reboot? [yes/N]", "yes")
	}

	flagStrategy := c.String("strategy")
	flagGroup := c.String("group")
	if flagGroup == "" {
		flagGroup = "web"
	}
	executeStackReboot(*stack, flagStrategy, flagGroup)
}

func executeStackReboot(stack cloud66.Stack, strategy string, group string) {
	fmt.Printf("Please wait while your server(s) are rebooted...\n")

	asyncId, err := startStackReboot(stack.Uid, strategy, group)
	if err != nil {
		printFatal(err.Error())
	}
	genericRes, err := endStackReboot(*asyncId, stack.Uid)
	if err != nil {
		printFatal(err.Error())
	}
	printGenericResponse(*genericRes)
	return
}

func startStackReboot(stackUid string, strategy string, group string) (*int, error) {
	asyncRes, err := client.StackReboot(stackUid, strategy, group)
	if err != nil {
		return nil, err
	}
	return &asyncRes.Id, err
}

func endStackReboot(asyncId int, stackUid string) (*cloud66.GenericResponse, error) {
	return client.WaitStackAsyncAction(asyncId, stackUid, 10*time.Second, 30*time.Minute, true)
}

package main

import (
	"fmt"

	"github.com/cloud66/cli"
)

// this is an alias for stacks redeploy command
var cmdRedeploy = &Command{
	Name:       "redeploy",
	Run:        runRedeploy,
	Build:      buildBasicCommand,
	NeedsStack: true,
	Short:      "this is a shortcut for stacks redeploy command",
}

func runRedeploy(c *cli.Context) {
	stack := mustStack(c)

	// confirmation is needed if the stack is production
	if stack.Environment == "production" && !c.Bool("y") {
		mustConfirm("This is a production stack. Proceed with deployment? [yes/N]", "yes")
	}
	result, err := client.RedeployStack(stack.Uid, c.String("git-ref"))
	if err != nil {
		printFatal(err.Error())
	} else {
		fmt.Println(result.Message)
	}
}

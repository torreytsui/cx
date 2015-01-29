package main

import (
	"fmt"

	"github.com/codegangsta/cli"
)

var cmdRedeploy = &Command{
	Name:       "redeploy",
	Run:        runRedeploy,
	Build:      buildBasicCommand,
	NeedsStack: true,
	Short:      "command shortcut: \"cx stacks help redeploy\"",
}

func runRedeploy(c *cli.Context) {
	stack := mustStack(c)

	// confirmation is needed if the stack is production
	if stack.Environment == "production" && !c.Bool("y") {
		mustConfirm("This is a production stack. Proceed with deployment? [yes/N]", "yes")
	}
	result, err := client.RedeployStack(stack.Uid, c.String("git-ref"), c.String("services"))
	if err != nil {
		printFatal(err.Error())
	} else {
		fmt.Println(result.Message)
	}
}

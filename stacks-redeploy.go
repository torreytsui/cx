package main

import (
	"fmt"

	"github.com/cloud66/cli"
)

// this is an alias for stacks redeploy command
var cmdRedeploy = &Command{
	Name:  "redeploy",
	Run:   runRedeploy,
	Build: buildBasicCommand,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "y",
			Usage: "answer yes to confirmations",
		},
		cli.StringFlag{
			Name:  "git-ref",
			Usage: "git reference",
		},
		cli.StringFlag{
			Name:  "services",
			Usage: "comma separated list of services to include in the deploy",
		},
	},

	NeedsStack: true,
	Short:      "An alias for 'stacks redeploy' command",
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

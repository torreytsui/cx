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
		cli.BoolFlag{
			Name:  "listen",
			Usage: "show stack deployment progress and log output",
		},
		cli.StringFlag{
			Name:  "git-ref",
			Usage: "[classic stacks] git reference",
		},
		cli.StringSliceFlag{
			Name:  "service",
			Usage: "[docker stacks] service name (and optional colon separated reference) to include in the deploy. Repeatable for multiple services",
			Value: &cli.StringSlice{},
		},
	},

	NeedsStack: true,
	NeedsOrg:   false,
	Short:      "An alias for 'stacks redeploy' command",
}

func runRedeploy(c *cli.Context) {
	stack := mustStack(c)

	// confirmation is needed if the stack is production
	if stack.Environment == "production" && !c.Bool("y") {
		mustConfirm("This is a production stack. Proceed with deployment? [yes/N]", "yes")
	}

	if len(c.StringSlice("service")) > 0 {
		fmt.Printf("Deploying service(s): ")
		for i, service := range c.StringSlice("service") {
			if i > 0 {
				fmt.Printf(", ")
			}
			fmt.Printf(service)
		}
		fmt.Printf("\n")
	}

	result, err := client.RedeployStack(stack.Uid, c.String("git-ref"), c.StringSlice("service"))
	must(err)

	if !c.Bool("listen") || result.Queued {
		// its queued - just message and exit
		fmt.Println(result.Message)
	} else {
		// tail the logs
		go StartListen(stack)

		stack, err = WaitStackBuild(stack.Uid, false)
		must(err)

		if stack.HealthCode == 2 || stack.HealthCode == 4 || stack.StatusCode == 2 || stack.StatusCode == 7 {
			fmt.Println("Completed with some errors!")
		} else {
			fmt.Println("Completed successfully!")
		}
	}
}

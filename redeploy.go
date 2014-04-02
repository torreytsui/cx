package main

import (
	"fmt"
	"os"
)

var cmdRedeploy = &Command{
	Run:        runRedeploy,
	Usage:      "redeploy",
	NeedsStack: true,
	Category:   "stack",
	Short:      "redeploys stack",
	Long: `Enqueues redeployment of the stack.
  If the stack is already building, another build will be enqueued and performed immediately
  after the current one is finished
`,
}

func runRedeploy(cmd *Command, args []string) {
	if len(args) != 0 {
		cmd.printUsage()
		os.Exit(2)
	}
	stack := mustStack()
	result, err := client.RedeployStack(stack.Uid)
	if err != nil {
		printFatal(err.Error())
	} else {
		fmt.Println(result.Message)
	}
}

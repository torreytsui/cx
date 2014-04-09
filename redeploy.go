package main

import (
	"fmt"
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
	stack := mustStack()
	result, err := client.RedeployStack(stack.Uid)
	if err != nil {
		printFatal(err.Error())
	} else {
		fmt.Println(result.Message)
	}
}

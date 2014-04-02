package main

import (
	"fmt"
)

var cmdRestart = &Command{
	Run:        runRestart,
	Usage:      "restart",
	NeedsStack: true,
	Category:   "stack",
	Short:      "restarts the stack.",
	Long: `This will send a restart method to all stack components. This means different things for different components.
For a web server, it means a restart of nginx. For an application server, this might be a restart of the workers like Unicorn.
For more information on restart command, please refer to help.cloud66.com
`,
}

func runRestart(cmd *Command, args []string) {
	stack := mustStack()
	result, err := client.RestartStack(stack.Uid)
	if err != nil {
		printFatal(err.Error())
	} else {
		fmt.Println(result.Message)
	}
}

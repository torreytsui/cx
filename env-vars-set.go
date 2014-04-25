package main

import (
	"fmt"
	"os"
)

var cmdEnvVarsSet = &Command{
	Run:        runEnvVarsSet,
	Usage:      "env-vars-set <setting> <value>",
	NeedsStack: true,
	Category:   "stack",
	Short:      "sets the value of an environment variable on a stack",
	Long: `This sets and applies the value of an environment variable on a stack.   
  This work happens in the background, therefore this command will return immediately after the operation has started.
  Warning! Applying environment variable changes to your stack will result in all your stack environment variables
  being sent to your stack servers, and your processes being restarted immediately. 

Examples:

    $ cx env-var-set -s mystack FIRST_VAR 123
    $ cx env-var-set -s mystack SECOND_ONE 'this value has a space in it'
`,
}

func runEnvVarsSet(cmd *Command, args []string) {
	if len(args) != 2 {
		cmd.printUsage()
		os.Exit(2)
	}

	key := args[0]
	value := args[1]

	stack := mustStack()
	result, err := client.StackEnvVarsSet(stack.Uid, key, value)
	if err != nil {
		printFatal(err.Error())
	} else {
		fmt.Println(result.Message)
	}
	return		
}	


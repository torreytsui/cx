// +build ignore

package main

import (
	"fmt"
)

var cmdRedeploy = &Command{
	Run:        runRedeploy,
	Usage:      "redeploy [-y] [--git-ref <git_ref>]",
	NeedsStack: true,
	Category:   "stack",
	Short:      "redeploys stack",
	Long: `Enqueues redeployment of the stack.
If the stack is already building, another build will be enqueued and performed immediately
after the current one is finished.

-y answers yes to confirmation question if the stack is production.
--git-ref will redeploy the specific git reference (branch, tag, hash)
`,
}

var flagConfirmation bool
var flagGitRef string

func init() {
	cmdRedeploy.Flag.BoolVar(&flagConfirmation, "y", false, "answer yes")
	cmdRedeploy.Flag.StringVar(&flagGitRef, "git-ref", "", "git reference")
}

func runRedeploy(cmd *Command, args []string) {
	stack := mustStack()

	// confirmation is needed if the stack is production
	if stack.Environment == "production" && !flagConfirmation {
		mustConfirm("This is a production stack. Proceed with deployment? [yes/N]", "yes")
	}
	result, err := client.RedeployStack(stack.Uid, flagGitRef)
	if err != nil {
		printFatal(err.Error())
	} else {
		fmt.Println(result.Message)
	}
}

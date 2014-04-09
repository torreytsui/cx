package main

import (
	"fmt"
	"os"
)

var cmdClearCaches = &Command{
	Run:        runClearCaches,
	Usage:      "clear-caches",
	NeedsStack: true,
	Category:   "stack",
	Short:      "clears all existing stack code caches",
	Long: `Clears all existing code caches.
  For improved performance, volatile code caches exist for your stack.
  It is possible for a those volatile caches to become invalid if you switch branches, change git repository URL, or rebase or force a commit. 
  Since switching branch or changing git repository URL is done via the Cloud 66 interface, your volatile caches will automatically be purged.
  However, rebasing or forcing a commit doesn't have any association with Cloud 66, so this command can be used to purge the exising volatile caches.
`,
}

func runClearCaches(cmd *Command, args []string) {
	if len(args) != 0 {
		cmd.printUsage()
		os.Exit(2)
	}
	stack := mustStack()
	result, err := client.ClearCachesStack(stack.Uid)
	if err != nil {
		printFatal(err.Error())
	} else {
		fmt.Println(result.Message)
	}
}

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/cloud66/cloud66"
)

var cmdSlaveResync = &Command{
	Run:        runSlaveResync,
	Usage:      "slave-resync [--db-type <db-type>] <server name>|<server ip>",
	NeedsStack: true,
	Category:   "stack",
	Short:      "re-syncs the specified slave database server with its master database server",
	Long: `Re-syncs the specified slave database server with its master database server.

From time-to-time your slave db server might go out of sync with its master. This action attempts to re-sync your specified slave server.
This can happen depending on many factors (such as DB size, frequency of change, networking between servers etc)

Examples:
$ slave-promote -s 'my stack name' postgresql_slave_name
$ slave-promote -s 'my stack name' --db-type=postgresql pg_slave1
`,
}

func runSlaveResync(cmd *Command, args []string) {
	stack := mustStack()

	if len(args) < 1 {
		cmd.printUsage()
		os.Exit(2)
	}

	// get the server
	serverName := args[0]

	servers, err := client.Servers(stack.Uid)
	if err != nil {
		printFatal(err.Error())
	}

	server, err := findServer(servers, serverName)
	if err != nil {
		printFatal(err.Error())
	}

	if server == nil {
		printFatal("Server '" + serverName + "' not found")
	}

	fmt.Printf("Server: %s\n", server.Name)
	currenttime := time.Now().Local()
	fmt.Printf("Started: %s\n", currenttime.Format("2006-01-02 15:04:05"))

	asyncId, err := startSlaveResync(stack.Uid, server.Uid, &flagDbType)
	if err != nil {
		printFatal(err.Error())
	}
	genericRes, err := endSlaveResync(*asyncId, stack.Uid)
	if err != nil {
		printFatal(err.Error())
	}
	printGenericResponse(*genericRes)
}

func startSlaveResync(stackUid string, serverUid string, dbType *string) (*int, error) {
	asyncRes, err := client.InvokeDbStackAction(stackUid, serverUid, dbType, "resync_slave_db")
	if err != nil {
		return nil, err
	}
	return &asyncRes.Id, err
}

func endSlaveResync(asyncId int, stackUid string) (*cloud66.GenericResponse, error) {
	return client.WaitStackAsyncAction(asyncId, stackUid, 5*time.Second, 20*time.Minute, true)
}

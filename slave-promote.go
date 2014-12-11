package main

import (
	"fmt"
	"os"
	"time"

	"github.com/cloud66/cloud66"
)

var cmdSlavePromote = &Command{
	Run:        runSlavePromote,
	Usage:      "slave-promote [--db-type <db-type>] <server name>|<server ip>",
	NeedsStack: true,
	Category:   "stack",
	Short:      "promotes the specified slave database server to a standalone master",
	Long: `Promotes the specified slave database server to a standalone master.

The slave will be reconfigured as the new standalone DB. The existing master and other existing slaves will be orphaned, and will need to be removed, after which you can scale them up again.

WARNING: This action could result in application downtime, it is advisable to choose a non-busy time to perform this action, and to place your stack in maintenance mode.
In the case of any servers not being accessible during this time, those servers will remain unchanged. It is therefore important to stop/shutdown those servers in this case.
(or to manually stop the DB service on those servers) as having multiple masters in a cluster could cause problems throughout the cluster.

Examples:
$ slave-promote -s 'my stack name' redis_slave_name
$ slave-promote -s 'my stack name' --db-type=postgresql pg_slave1
`,
}

func runSlavePromote(cmd *Command, args []string) {
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

	asyncId, err := startSlavePromote(stack.Uid, server.Uid, &flagDbType)
	if err != nil {
		printFatal(err.Error())
	}
	genericRes, err := endSlavePromote(*asyncId, stack.Uid)
	if err != nil {
		printFatal(err.Error())
	}
	printGenericResponse(*genericRes)
}

func startSlavePromote(stackUid string, serverUid string, dbType *string) (*int, error) {
	asyncRes, err := client.InvokeDbStackAction(stackUid, serverUid, dbType, "promote_slave_db")
	if err != nil {
		return nil, err
	}
	return &asyncRes.Id, err
}

func endSlavePromote(asyncId int, stackUid string) (*cloud66.GenericResponse, error) {
	return client.WaitStackAsyncAction(asyncId, stackUid, 5*time.Second, 20*time.Minute, true)
}

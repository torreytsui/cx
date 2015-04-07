package main

import (
	"fmt"
	"os"
	"time"

	"github.com/cloud66/cloud66"

	"github.com/cloud66/cli"
)

func runSlaveResync(c *cli.Context) {
	stack := mustStack(c)

	if len(c.Args()) < 1 {
		cli.ShowSubcommandHelp(c)
		os.Exit(2)
	}

	// get the server
	serverName := c.Args()[0]
	flagDbType := c.String("dbtype")

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
	return client.WaitStackAsyncAction(asyncId, stackUid, 5*time.Second, 2*time.Hour, true)
}

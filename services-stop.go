package main

import (
	"fmt"
	"os"
	"time"

	"github.com/cloud66/cloud66"

	"github.com/codegangsta/cli"
)

func runServiceStop(c *cli.Context) {
	if len(c.Args()) != 1 {
		cli.ShowSubcommandHelp(c)
		os.Exit(2)
	}

	stack := mustStack(c)
	serviceName := c.Args()[0]

	flagServer := c.String("server")

	var serverUid *string
	if flagServer == "" {
		serverUid = nil
	} else {
		servers, err := client.Servers(stack.Uid)
		if err != nil {
			printFatal(err.Error())
		}
		server, err := findServer(servers, flagServer)
		if err != nil {
			printFatal(err.Error())
		}
		if server == nil {
			printFatal("Server '" + flagServer + "' not found")
		}
		if !server.HasRole("docker") {
			printFatal("Server '" + flagServer + "' can not host containers")
		}
		fmt.Printf("Server: %s\n", server.Name)
		serverUid = &server.Uid
	}

	asyncId, err := startServiceStop(stack.Uid, serviceName, serverUid)
	if err != nil {
		printFatal(err.Error())
	}
	genericRes, err := endServiceStop(*asyncId, stack.Uid)
	if err != nil {
		printFatal(err.Error())
	}
	printGenericResponse(*genericRes)
	return
}

func startServiceStop(stackUid string, serviceName string, serverUid *string) (*int, error) {
	asyncRes, err := client.StopService(stackUid, serviceName, serverUid)
	if err != nil {
		return nil, err
	}
	return &asyncRes.Id, err
}

func endServiceStop(asyncId int, stackUid string) (*cloud66.GenericResponse, error) {
	return client.WaitStackAsyncAction(asyncId, stackUid, 5*time.Second, 20*time.Minute, true)
}

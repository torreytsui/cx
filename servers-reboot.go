package main

import (
	"fmt"
	"os"
	"time"

	"github.com/cloud66/cloud66"

	"github.com/cloud66/cli"
)

func runServerReboot(c *cli.Context) {
	fmt.Println(c.Args())
	stack := mustStack(c)

	if len(c.Args()) != 1 {
		cli.ShowSubcommandHelp(c)
		os.Exit(2)
	}

	serverName := c.String("server")
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

	// confirmation is needed if the stack is production
	if !c.Bool("y") {
		mustConfirm("This operation will reboot your server during which time your server will not be available. Proceed with reboot? [yes/N]", "yes")
	}
	executeServerReboot(*stack, *server)
}

func executeServerReboot(stack cloud66.Stack, server cloud66.Server) {
	fmt.Printf("Please wait while your server is rebooted...\n")

	asyncId, err := startServerReboot(stack.Uid, server.Uid)
	if err != nil {
		printFatal(err.Error())
	}
	genericRes, err := endServerReboot(*asyncId, stack.Uid)
	if err != nil {
		printFatal(err.Error())
	}
	printGenericResponse(*genericRes)
	return
}

func startServerReboot(stackUid string, serverUid string) (*int, error) {
	asyncRes, err := client.ServerReboot(stackUid, serverUid)
	if err != nil {
		return nil, err
	}
	return &asyncRes.Id, err
}

func endServerReboot(asyncId int, stackUid string) (*cloud66.GenericResponse, error) {
	return client.WaitStackAsyncAction(asyncId, stackUid, 5*time.Second, 20*time.Minute, true)
}

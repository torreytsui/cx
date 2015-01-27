package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cloud66/cloud66"

	"github.com/codegangsta/cli"
)

func runServerSet(c *cli.Context) {
	fmt.Println(c.Args())
	stack := mustStack(c)

	if len(c.Args()) != 1 {
		cli.ShowSubcommandHelp(c)
		os.Exit(2)
	}

	// get the server
	args := c.Args()

	serverName := c.String("server")

	// filter out the server name
	kvs := args[0]
	kva := strings.Split(kvs, "=")
	if len(kva) != 2 {
		cli.ShowSubcommandHelp(c)
		os.Exit(2)
	}
	key := kva[0]
	value := kva[1]

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

	executeServerSet(*stack, *server, c, key, value)
}

func executeServerSet(stack cloud66.Stack, server cloud66.Server, c *cli.Context, key string, value string) {
	settings, err := client.ServerSettings(stack.Uid, server.Uid)
	must(err)

	// check to see if it's a valid setting
	for _, i := range settings {
		if key == i.Key {
			// yup. it's a good one
			fmt.Printf("Please wait while your setting is applied...\n")

			asyncId, err := startServerSet(stack.Uid, server.Uid, key, value)
			if err != nil {
				printFatal(err.Error())
			}
			genericRes, err := endServerSet(*asyncId, stack.Uid)
			if err != nil {
				printFatal(err.Error())
			}
			printGenericResponse(*genericRes)

			return
		}
	}

	printFatal(key + " is not a valid setting or does not apply to this server")
}

func startServerSet(stackUid string, serverUid string, key string, value string) (*int, error) {
	asyncRes, err := client.ServerSet(stackUid, serverUid, key, value)
	if err != nil {
		return nil, err
	}
	return &asyncRes.Id, err
}

func endServerSet(asyncId int, stackUid string) (*cloud66.GenericResponse, error) {
	return client.WaitStackAsyncAction(asyncId, stackUid, 5*time.Second, 20*time.Minute, true)
}

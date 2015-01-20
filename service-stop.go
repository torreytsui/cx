package main

import (
	"fmt"
	"os"
	"time"

	"github.com/cloud66/cloud66"
)

var cmdServiceStop = &Command{
	Run:        runServiceStop,
	Usage:      "service-stop [--server <server name>|<server ip>|<server role>] <service name>",
	NeedsStack: true,
	Category:   "stack",
	Short:      "stops all the containers from the given service",
	Long: `Stops all the containers from the given service.
The list of available stack services can be obtained through the 'services' command.
If the server is provided it will only act on the specified server.

Examples:
$ cx service-stop -s mystack my_web_service
$ cx service-stop -s mystack a_backend_service
$ cx service-stop -s mystack --server my_server my_web_service
`,
}

func runServiceStop(cmd *Command, args []string) {
	if len(args) != 1 {
		cmd.printUsage()
		os.Exit(2)
	}

	stack := mustStack()
	serviceName := args[0]

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

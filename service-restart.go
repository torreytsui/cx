// +build ignore

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/cloud66/cloud66"
)

var cmdServiceRestart = &Command{
	Run:        runServiceRestart,
	Usage:      "service-restart <service name> [--server <server name>|<server ip>|<server role>]",
	NeedsStack: true,
	Category:   "stack",
	Short:      "restarts all the containers from the given service",
	Long: `Restarts all the containers from the given service.
The list of available stack services can be obtained through the 'services' command.
If the server is provided it will only act on the specified server.

Examples:
$ cx service-restart -s mystack my_web_service
$ cx service-restart -s mystack a_backend_service
$ cx service-restart -s mystack --server my_server my_web_service
`,
}

func runServiceRestart(cmd *Command, args []string) {
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
			printFatal("Server '" + flagServer + "' is not a docker server")
		}
		fmt.Printf("Server: %s\n", server.Name)
		serverUid = &server.Uid
	}

	asyncId, err := startServiceRestart(stack.Uid, serviceName, serverUid)
	if err != nil {
		printFatal(err.Error())
	}
	genericRes, err := endServiceRestart(*asyncId, stack.Uid)
	if err != nil {
		printFatal(err.Error())
	}
	printGenericResponse(*genericRes)
	return
}

func startServiceRestart(stackUid string, serviceName string, serverUid *string) (*int, error) {
	asyncRes, err := client.InvokeStackServiceAction(stackUid, serviceName, serverUid, "service_restart")
	if err != nil {
		return nil, err
	}
	return &asyncRes.Id, err
}

func endServiceRestart(asyncId int, stackUid string) (*cloud66.GenericResponse, error) {
	return client.WaitStackAsyncAction(asyncId, stackUid, 5*time.Second, 20*time.Minute, true)
}

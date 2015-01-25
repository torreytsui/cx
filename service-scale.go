// +build ignore

package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cloud66/cloud66"
)

var cmdServiceStart = &Command{
	Run:        runServiceScale,
	Usage:      "service-scale <service name> [--server <server name>|<server ip>|<server role>] <count>",
	NeedsStack: true,
	Category:   "stack",
	Short:      "starts containers from the given service",
	Long: `Starts <count> containers from the given service.
<count> can be an absolute value like "2" or a relative value like "+2" or "-3" etc.
If server is provided, will start <count> containers on that server.
If server is not provided, will start <count> containers on every server.

Examples:
$ cx service-scale -s mystack my_web_service 1
$ cx service-scale -s mystack a_backend_service --server backend +5
$ cx service-scale -s mystack a_backend_service -2
`,
}

func runServiceScale(cmd *Command, args []string) {
	if len(args) != 2 {
		cmd.printUsage()
		os.Exit(2)
	}

	stack := mustStack()
	serviceName := args[0]
	count := args[1]

	// fetch servers info
	servers, err := client.Servers(stack.Uid)
	must(err)

	if flagServer != "" {
		server, err := findServer(servers, flagServer)
		must(err)
		if server == nil {
			printFatal("Server '" + flagServer + "' not found")
		}
		if !server.HasRole("docker") {
			printFatal("Server '" + flagServer + "' can not host containers")
		}
		fmt.Printf("Server: %s\n", server.Name)
		// filter servers collection down
		servers = make([]cloud66.Server, 1)
		servers[0] = *server
	}

	// param for api call
	serverCountDesired := make(map[string]int)

	var absoluteCount int
	if strings.ContainsAny(count, "+ & -") {

		// fetch service information for existing counts
		service, err := client.GetService(stack.Uid, serviceName, nil, nil)
		must(err)

		serverCountCurrent := service.ServerContainerCountMap()
		relativeCount, _ := strconv.Atoi(count)

		for _, server := range servers {
			if _, present := serverCountCurrent[server.Name]; present {
				serverCountDesired[server.Uid] = relativeCount + serverCountCurrent[server.Name]
			} else {
				serverCountDesired[server.Uid] = relativeCount
			}
		}

	} else {
		absoluteCount, _ = strconv.Atoi(count)
		for _, server := range servers {
			serverCountDesired[server.Uid] = absoluteCount
		}
	}

	// validate non < 0
	for serverUid, count := range serverCountDesired {
		if count < 0 {
			serverCountDesired[serverUid] = 0
		}
	}

	fmt.Println("Scaling your '" + serviceName + "' service")

	asyncId, err := startServiceScale(stack.Uid, serviceName, serverCountDesired)
	must(err)
	genericRes, err := endServiceScale(*asyncId, stack.Uid)
	must(err)
	printGenericResponse(*genericRes)
	return
}

func startServiceScale(stackUid string, serviceName string, serverCount map[string]int) (*int, error) {
	asyncRes, err := client.ScaleService(stackUid, serviceName, serverCount)
	must(err)
	return &asyncRes.Id, err
}

func endServiceScale(asyncId int, stackUid string) (*cloud66.GenericResponse, error) {
	return client.WaitStackAsyncAction(asyncId, stackUid, 5*time.Second, 20*time.Minute, true)
}

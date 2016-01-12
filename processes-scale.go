package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cloud66/cli"
	"github.com/cloud66/cloud66"
)

func runProcessScale(c *cli.Context) {
	if len(c.Args()) != 2 {
		cli.ShowSubcommandHelp(c)
		os.Exit(2)
	}

	stack := mustStack(c)
	if stack.Framework == "docker" {
		printFatal("Stack '" + stack.Name + "' is a docker stack; 'cx processes' is not supported, use 'cx services' instead")
	}

	processName := c.Args()[0]
	count := c.Args()[1]

	flagServer := c.String("server")
	// fetch servers info
	servers, err := client.Servers(stack.Uid)
	must(err)

	if flagServer != "" {
		server, err := findServer(servers, flagServer)
		must(err)
		if server == nil {
			printFatal("Server '" + flagServer + "' not found")
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
		service, err := client.GetService(stack.Uid, processName, nil, nil)
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

	fmt.Println("Scaling your '" + processName + "' service")

	var asyncId *int
	if flagGroup != "" {
		var groupMap = make(map[string]int)
		groupMap["web"] = absoluteCount
		asyncId, err = startServiceScaleByGroup(stack.Uid, processName, groupMap)
	} else {
		asyncId, err = startServiceScale(stack.Uid, processName, serverCountDesired)
	}

	must(err)
	genericRes, err := endServiceScale(*asyncId, stack.Uid)
	must(err)
	printGenericResponse(*genericRes)
	return
}

func startServiceScale(stackUid string, processName string, serverCount map[string]int) (*int, error) {
	asyncRes, err := client.ScaleService(stackUid, processName, serverCount)
	must(err)
	return &asyncRes.Id, err
}

func startServiceScaleByGroup(stackUid string, processName string, groupCount map[string]int) (*int, error) {
	asyncRes, err := client.ScaleServiceByGroup(stackUid, processName, groupCount)
	must(err)
	return &asyncRes.Id, err
}

func endServiceScale(asyncId int, stackUid string) (*cloud66.GenericResponse, error) {
	return client.WaitStackAsyncAction(asyncId, stackUid, 5*time.Second, 20*time.Minute, true)
}

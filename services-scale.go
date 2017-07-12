package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cloud66/cloud66"

	"github.com/cloud66/cli"
)

func runServiceScale(c *cli.Context) {
	if len(c.Args()) != 2 {
		cli.ShowSubcommandHelp(c)
		os.Exit(2)
	}

	stack := mustStack(c)
	serviceName := c.Args()[0]

	count := c.Args()[1]
	count = strings.Replace(count, "[", "", -1)
	count = strings.Replace(count, "]", "", -1)
	count = strings.Replace(count, " ", "", -1)

	flagServer := c.String("server")
	flagGroup := c.String("group")

	// fetch servers info
	servers, err := client.Servers(stack.Uid)
	must(err)

	if flagServer != "" {
		server, err := findServer(servers, flagServer)
		must(err)
		if server == nil {
			printFatal("Server '" + flagServer + "' not found")
		}
		if !server.HasRole("docker") && !server.HasRole("kubes") {
			printFatal("Server '" + flagServer + "' can not host containers")
		}
		fmt.Printf("Server: %s\n", server.Name)
		// filter servers collection down
		servers = make([]cloud66.Server, 1)
		servers[0] = *server
	}

	if flagGroup != "" {
		if flagGroup != "web" {
			printFatal("Only web group is supported at the moment")
		}
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

	var asyncId *int
	if flagGroup != "" {
		var groupMap = make(map[string]int)
		groupMap["web"] = absoluteCount
		asyncId, err = startServiceScaleByGroup(stack.Uid, serviceName, groupMap)
	} else {
		asyncId, err = startServiceScale(stack.Uid, serviceName, serverCountDesired)
	}

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

func startServiceScaleByGroup(stackUid string, serviceName string, groupCount map[string]int) (*int, error) {
	asyncRes, err := client.ScaleServiceByGroup(stackUid, serviceName, groupCount)
	must(err)
	return &asyncRes.Id, err
}

func endServiceScale(asyncId int, stackUid string) (*cloud66.GenericResponse, error) {
	return client.WaitStackAsyncAction(asyncId, stackUid, 5*time.Second, 20*time.Minute, true)
}

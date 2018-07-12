package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cloud66-oss/cloud66"

	"github.com/cloud66/cli"
)

func runServiceScale(c *cli.Context) {
	if len(c.Args()) != 2 {
		cli.ShowSubcommandHelp(c)
		os.Exit(2)
	}

	stack := mustStack(c)

	serviceName := c.Args()[0]
	service, err := client.GetService(stack.Uid, serviceName, nil, nil)
	must(err)

	count := c.Args()[1]
	count = strings.Replace(count, "[", "", -1)
	count = strings.Replace(count, "]", "", -1)
	count = strings.Replace(count, " ", "", -1)

	var absoluteCount int
	if strings.ContainsAny(count, "+ & -") {
		var serviceCountCurrent int
		if stack.Backend == "kubernetes" {
			serviceCountCurrent = service.DesiredCount
		} else {
			serviceCountCurrent = len(service.Containers)
		}

		relativeCount, err := strconv.Atoi(count)
		if err != nil {
			printFatal("Could not parse the <count> argument provided (" + count + ")")
		}

		absoluteCount = serviceCountCurrent + relativeCount
	} else {
		parsedAbsoluteCount, err := strconv.Atoi(count)
		if err != nil {
			printFatal("Could not parse the <count> argument provided (" + count + ")")
		}

		absoluteCount = parsedAbsoluteCount
	}

	if absoluteCount < 0 {
		printWarning("With the <count> argument provided, the resulting amount of containers would be less than 0. Will scale to 0 instead.")
		absoluteCount = 0
	}

	fmt.Println("Scaling your '" + serviceName + "' service to " + strconv.Itoa(absoluteCount) + " containers.")

	// param for api call
	groupMap := make(map[string]int)
	groupMap["web"] = absoluteCount

	var asyncId *int
	asyncId, err = startServiceScaleByGroup(stack.Uid, serviceName, groupMap)
	must(err)

	genericRes, err := endServiceScale(*asyncId, stack.Uid)
	must(err)

	printGenericResponse(*genericRes)

	return
}

func startServiceScaleByGroup(stackUid string, serviceName string, groupCount map[string]int) (*int, error) {
	asyncRes, err := client.ScaleServiceByGroup(stackUid, serviceName, groupCount)
	must(err)
	return &asyncRes.Id, err
}

func endServiceScale(asyncId int, stackUid string) (*cloud66.GenericResponse, error) {
	return client.WaitStackAsyncAction(asyncId, stackUid, 5*time.Second, 20*time.Minute, true)
}

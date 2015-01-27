package main

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/cloud66/cloud66"

	"github.com/codegangsta/cli"
)

func runServerSettings(c *cli.Context) {
	stack := mustStack(c)

	// get the server
	serverName := c.String("server")
	if len(serverName) == 0 {
		cli.ShowSubcommandHelp(c)
		os.Exit(2)
	}

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

	getServerSettings(*stack, *server, c.Args())
}

func getServerSettings(stack cloud66.Stack, server cloud66.Server, settingNames []string) {
	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()
	var settings []cloud66.StackSetting
	var err error

	settings, err = client.ServerSettings(stack.Uid, server.Uid)
	must(err)

	sort.Strings(settingNames)
	if len(settingNames) == 0 {
		printSettingList(w, settings)
	} else {
		// filter out the unwanted settings
		var filteredSettings []cloud66.StackSetting
		for _, i := range settings {
			sorted := sort.SearchStrings(settingNames, i.Key)
			if sorted < len(settingNames) && settingNames[sorted] == i.Key {
				filteredSettings = append(filteredSettings, i)
			}
		}

		printSettingList(w, filteredSettings)
	}
}

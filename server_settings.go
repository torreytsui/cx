package main

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/cloud66/cx/cloud66"
)

var cmdServerSettings = &Command{
	Run:        runServerSettings,
	Usage:      "server-settings <server name>|<server ip>|<server role> [settings]",
	NeedsStack: true,
	Category:   "stack",
	Short:      "lists server settings",
	Long: `Lists all the settings applicable to the given server.
  It also shows the key, value and the readonly flag for each setting.
  Settings can be a list of multiple settings as separate parameters.
  To change each server setting, use the server-set command.

Examples:

    $ cx server-settings -s mystack lion
    server.name         lion                                                       false

    $ cx server-settings -s mystack db server-name
    server.name         tiger                                                      false
`,
}

func runServerSettings(cmd *Command, args []string) {

	stack := mustStack()

	if len(args) < 1 {
		cmd.printUsage()
		os.Exit(2)
	}

	// get the server
	serverName := args[0]

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

	getServerSettings(*stack, *server, args)
}

func getServerSettings(stack cloud66.Stack, server cloud66.Server, settingNames []string) {

	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()
	var settings []cloud66.StackSetting
	var err error

	settings, err = client.ServerSettings(server.Uid)
	must(err)

	// filter out the server name
	settingNames = settingNames[1:len(settingNames)]

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

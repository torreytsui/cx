package main

import (
	"fmt"
)

var cmdOpen = &Command{
	Run:        runOpen,
	Usage:      "open [-v <server name>|<server ip>]",
	NeedsStack: true,
	Category:   "stack",
	Short:      "opens the web browser to visit the app served by the stack",
	Long: `This opens the client web browser to visit the app servers by the stack. This could be the web page
  specifically served by one server or the load balancer.

  If no server is specified, the command opens the page served by the stack load balancer or first web server.

  -v  Specific server to visit. Could be the name or IP of the server. Partial names are accepted and are case insensitive.

  Examples:

    $ cx open
    $ cx open -s lion
  `,
}

var (
	flagServer string
)

func init() {
	cmdOpen.Flag.StringVar(&flagServer, "v", "", "server to connect to")
}

func runOpen(cmd *Command, args []string) {
	stack := mustStack()

	var toOpen string
	// are we connecting to a server?
	if flagServer != "" {
		// find the server
		// get stack servers
		servers, err := client.Servers(stack.Uid)
		if err != nil {
			printFatal(err.Error())
		}
		server, err := findServer(servers, flagServer)
		if err != nil {
			printFatal(err.Error())
		}
		if server == nil {
			printFatal("Server not found")
		}

		toOpen = "http://" + server.DnsRecord

	} else {
		// is the stack load balanced?
		if stack.HasLoadBalancer {
			toOpen = "http://" + stack.Fqdn
		} else {
			// use the first web server
			servers, err := client.Servers(stack.Uid)
			if err != nil {
				printFatal(err.Error())
			}

			toOpen = "http://" + servers[0].DnsRecord
		}
	}

	// open server's fqdn
	fmt.Printf("Openning %s\n", toOpen)
	err := openURL(toOpen)
	if err != nil {
		printFatal(err.Error())
	}
}

package main

import (
	"fmt"
	"os"
)

var cmdOpen = &Command{
	Run:        runOpen,
	Usage:      "open [<server name>|<server ip>|<server role>]",
	NeedsStack: true,
	Category:   "stack",
	Short:      "opens the web browser to visit the app served by the stack",
	Long: `This opens the client web browser to visit the app servers by the stack. This could be the web page
  specifically served by one server or the load balancer.

  If no server is specified, the command opens the page served by the stack load balancer or first web server.
  Alternatively you can specify the name or IP of the server. Partial names are accepted and are case insensitive.

  Examples:

    $ cx open
    $ cx open lion
    $ cx open -s mystack
    $ cx open -s mystack lion
  `,
}

// var (
// 	flagServer string
// )

// func init() {
// 	cmdOpen.Flag.StringVar(&flagServer, "v", "", "server to connect to")
// }

func runOpen(cmd *Command, args []string) {
	stack := mustStack()

	if len(args) > 1 {
		cmd.printUsage()
		os.Exit(2)
	}

	var toOpen string
	// are we connecting to a server?
	if len(args) == 1 {

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
			fmt.Printf("Server: %s\n", servers[0].Name)		
			toOpen = "http://" + servers[0].DnsRecord
		}
	}

	// open server's fqdn
	fmt.Printf("Opening %s\n", toOpen)
	err := openURL(toOpen)
	if err != nil {
		printFatal(err.Error())
	}
}

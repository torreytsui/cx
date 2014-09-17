package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/cloud66/cloud66"
)

var cmdServices = &Command{
	Run:        runServices,
	Usage:      "services [--server <server name>|<server ip>|<server role>]",
	NeedsStack: true,
	Category:   "stack",
	Short:      "lists all the services of a stack (or server)",
	Long: `List all the services and running containers of a stack or a server.

  Examples:

  $ cx services -s mystack
  $ cx services -s mystack --server orca
`,
}

func runServices(cmd *Command, args []string) {
	if len(args) > 0 {
		cmd.printUsage()
		os.Exit(2)
	}

	stack := mustStack()
	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()

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

	var (
		services []cloud66.Service
		err      error
	)

	if flagServiceName == "" {
		services, err = client.GetServices(stack.Uid, serverUid)
		must(err)
	} else {
		service, err := client.GetService(stack.Uid, flagServiceName, serverUid)
		must(err)
		if service == nil {
			printFatal("Service '" + flagServiceName + "' not found on specified stack")
		} else {
			services = make([]cloud66.Service, 1)
			services[0] = *service
		}
	}
	printServicesList(w, services)
}

func printServicesList(w io.Writer, services []cloud66.Service) {
	listRec(w,
		"SERVICE NAME",
		"SERVER",
		"COUNT",
	)

	sort.Sort(ServiceByNameServer(services))
	for _, a := range services {
		listService(w, a)
	}
}

func listService(w io.Writer, a cloud66.Service) {
	if len(a.Containers) != 0 {
		for serverName, count := range a.ServerContainerCountMap() {
			listRec(w,
				a.Name,
				serverName,
				count,
			)
		}
	} else if flagServer == "" {
		listRec(w,
			a.Name,
			"n/a",
			"0",
		)
	}

}

type ServiceByNameServer []cloud66.Service

func (a ServiceByNameServer) Len() int           { return len(a) }
func (a ServiceByNameServer) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ServiceByNameServer) Less(i, j int) bool { return a[i].Name < a[j].Name }

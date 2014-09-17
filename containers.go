package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"text/tabwriter"

	"github.com/cloud66/cloud66"
)

var cmdContainers = &Command{
	Run:        runContainers,
	Usage:      "containers [--server <server name>|<server ip>|<server role>]",
	NeedsStack: true,
	Category:   "stack",
	Short:      "lists all the running containers of a stack (or server)",
	Long: `List all the running containers of a stack or a server.

Examples:
$ cx containers -s mystack
$ cx containers -s mystack --server orca
`,
}

func runContainers(cmd *Command, args []string) {
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
		fmt.Printf("Server: %s\n", server.Name)
		serverUid = &server.Uid
	}

	containers, err := client.GetContainers(stack.Uid, serverUid, &flagServiceName)
	must(err)

	printContainerList(w, containers)
}

func printContainerList(w io.Writer, containers []cloud66.Container) {
	listRec(w,
		"SERVICE",
		"SERVER",
		"CONTAINER ID",
		"IMAGE",
		"STARTED AT",
	)

	sort.Sort(containersByService(containers))
	for _, a := range containers {
		if a.Uid != "" {
			listContainer(w, a)
		}
	}
}

func listContainer(w io.Writer, a cloud66.Container) {
	t := a.StartedAt

	// fmt.Println(t.Format("20060102150405"))

	listRec(w,
		strings.ToLower(a.ServiceName),
		a.ServerName,
		a.Uid,
		a.Image,
		prettyTime{t},
	)
}

type containersByService []cloud66.Container

func (a containersByService) Len() int           { return len(a) }
func (a containersByService) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a containersByService) Less(i, j int) bool { return a[i].ServiceName < a[j].ServiceName }

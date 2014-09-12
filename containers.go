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
	Usage:      "containers",
	NeedsStack: true,
	Category:   "stack",
	Short:      "lists all the running containers of a stack (or server)",
	Long: `List all the running containers of a stack or a server.

  Examples:

  $ cx containers -s mystack
  $ cx containers -s mystack orca
`,
}

var (
	flagServer string
)

func init() {
	cmdContainers.Flag.StringVar(&flagServer, "server", "", "server filter")
}

func runContainers(cmd *Command, args []string) {
	stack := mustStack()
	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()

	if len(args) > 0 {
		cmd.printUsage()
		os.Exit(2)
	}

	var server *cloud66.Server
	if flagServer != "" {
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
	} else {
		server = nil
	}

	var (
		containers []cloud66.Container
		container  *cloud66.Container
		err        error
	)
	if server == nil {
		containers, err = client.GetContainers(stack.Uid)
		must(err)
	} else {
		container, err = client.GetContainer(stack.Uid, server.Uid)
		must(err)

		containers = make([]cloud66.Container, 1, 1)
		if container != nil {
			containers[0] = *container
		}
	}

	printContainerList(w, containers)
}

func printContainerList(w io.Writer, containers []cloud66.Container) {
	listRec(w,
		"CONTAINER ID",
		"SERVICE NAME",
		"SERVER UID",
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
		a.Uid,
		strings.ToLower(a.ServiceName),
		a.ServerUid,
		a.Image,
		prettyTime{t},
	)
}

type containersByService []cloud66.Container

func (a containersByService) Len() int           { return len(a) }
func (a containersByService) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a containersByService) Less(i, j int) bool { return a[i].ServiceName < a[j].ServiceName }

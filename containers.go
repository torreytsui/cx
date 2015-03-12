package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"text/tabwriter"

	"github.com/cloud66/cloud66"

	"github.com/cloud66/cli"
)

var cmdContainers = &Command{
	Name:       "containers",
	Build:      buildContainers,
	Short:      "commands to work with containers",
	NeedsStack: true,
}

func buildContainers() cli.Command {
	base := buildBasicCommand()

	base.Subcommands = []cli.Command{
		cli.Command{
			Name:   "list",
			Action: runContainers,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "server",
					Usage: "server to target",
				},
			},
			Usage: "lists all the running containers of a stack (or server)",
			Description: `List all the running containers of a stack or a server.

Examples:
$ cx containers list -s mystack
$ cx containers list -s mystack --server orca
`,
		},
		cli.Command{
			Name:   "stop",
			Action: runContainerStop,
			Usage:  "Stops a particular container on the given stack",
			Description: `Stops a particular container on the given stack by container Id.

Examples:
$ cx containers stop -s mystack 2844142cbfc064123777b6be765b3914e43a9e083afce4e4348b5979127c220c
`,
		},
		cli.Command{
			Name:   "restart",
			Action: runContainerRestart,
			Usage:  "Restarts a particular container on the given stack",
			Description: `Restarts a particular container on the given stack by container Id.
Examples:
$ cx containers restart -s mystack 2844142cbfc064123777b6be765b3914e43a9e083afce4e4348b5979127c220c
`,
		},
		cli.Command{
			Name:   "attach",
			Action: runContainerAttach,
			Usage:  "Attach to a container on the given stack",
			Description: `Attach to a container on the given stack by container Id.
Examples:
$ cx containers attach -s mystack 2844142cbfc064123777b6be765b3914e43a9e083afce4e4348b5979127c220c
`,
		},
	}
	return base
}

func runContainers(c *cli.Context) {
	stack := mustStack(c)
	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()

	flagServer := c.String("server")
	flagServiceName := c.String("service")

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
		"CONTAINER_NET_IP",
		"DOCKER_IP",
		"IMAGE",
		"STARTED AT",
		"HEALTH",
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

	listRec(w,
		strings.ToLower(a.ServiceName),
		a.ServerName,
		abbrev(a.Uid, 16),
		a.PrivateIP,
		a.DockerIP,
		a.Image,
		prettyTime{t},
		HealthText(a),
	)
}

type containersByService []cloud66.Container

func (a containersByService) Len() int           { return len(a) }
func (a containersByService) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a containersByService) Less(i, j int) bool { return a[i].ServiceName < a[j].ServiceName }

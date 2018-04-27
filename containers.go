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
	NeedsOrg:   false,
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
				cli.BoolFlag{
					Name:  "verbose",
					Usage: "Show more information about each container",
				},
			},
			Usage: "lists all the running containers of a stack (or server)",
			Description: `List all the running containers of a stack or a server. Optionally can truncate container Ids for easier reading.

Examples:
$ cx containers list -s mystack
$ cx containers list -s mystack --server orca
$ cx containers list -s mystack --verbose --server orca
`,
		},
		cli.Command{
			Name:   "stop",
			Action: runContainerStop,
			Usage:  "Stops a particular container on the given stack",
			Description: `Stops a particular container on the given stack by container Id or Name.

Examples:
$ cx containers stop -s mystack 2844142cbfc064123777b6be765b3914e43a9e083afce4e4348b5979127c220c
$ cx containers stop -s mystack 2844142c
$ cx containers stop -s mystack web.pro-active-quick-witted-dinosaur
$ cx containers stop -s mystack web
`,
		},
		cli.Command{
			Name:   "restart",
			Action: runContainerRestart,
			Usage:  "Restarts a particular container on the given stack",
			Description: `Restarts a particular container on the given stack by container Id or Name.

Examples:
$ cx containers stop -s mystack 2844142cbfc064123777b6be765b3914e43a9e083afce4e4348b5979127c220c
$ cx containers stop -s mystack 2844142c
$ cx containers stop -s mystack web.pro-active-quick-witted-dinosaur
$ cx containers stop -s mystack web
`,
		},
		cli.Command{
			Name:   "exec",
			Action: runContainerExec,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "cli-flags,  docker-flags",
					Usage: "specify cli flags",
				},
			},
			Usage: "Execute a command within the context of a running container",
			Description: `Execute a command within the context of a running container. The default cli-flags are for an interactive shell though they can be specified with the command.
   NOTE: the cli for Container v1 stacks is docker, but for Container v2 stacks it is kubectl, so be aware that the cli-flags will be different.

   Examples:
   $ cx containers exec -s mystack container_uid command
   $ cx containers exec -s mystack --cli-flags="--stdin=true --tty=true" container_uid /bin/bash
   $ cx container exec  -s mystack container_uid 'ls -al'
`,
		},
		cli.Command{
			Name:   "attach",
			Action: runContainerAttach,
			Usage:  "Attach to a container on the given stack",
			Description: `Attach to a container on the given stack by container Id.
Examples:
$ cx containers attach -s mystack 2844142c
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
	flagVerbose := c.Bool("verbose")

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

	printContainerList(w, containers, flagVerbose)
}

func printContainerList(w io.Writer, containers []cloud66.Container, flagVerbose bool) {
	if flagVerbose {
		listRec(w,
			"SERVICE",
			"SERVER",
			"NAME",
			"CONTAINER ID",
			"CONTAINER_NET_IP",
			"DOCKER_IP",
			"IMAGE",
			"STARTED AT",
			"HEALTH")
	} else {
		listRec(w,
			"SERVICE",
			"SERVER",
			"NAME",
			"CONTAINER ID",
			"CONTAINER_NET_IP",
			"DOCKER_IP",
			"STARTED AT",
			"HEALTH")
	}

	sort.Sort(containersByService(containers))
	for _, a := range containers {
		if a.Uid != "" {
			listContainer(w, a, flagVerbose)
		}
	}
}

func listContainer(w io.Writer, a cloud66.Container, flagVerbose bool) {
	t := a.StartedAt

	var containerId string
	if flagVerbose {
		containerId = a.Uid
	} else {
		containerId = abbrev(a.Uid, 16)
	}

	if flagVerbose {
		listRec(w,
			strings.ToLower(a.ServiceName),
			a.ServerName,
			a.Name,
			containerId,
			a.PrivateIP,
			a.DockerIP,
			a.Image,
			prettyTime{t},
			HealthText(a))
	} else {
		listRec(w,
			strings.ToLower(a.ServiceName),
			a.ServerName,
			a.Name,
			containerId,
			a.PrivateIP,
			a.DockerIP,
			prettyTime{t},
			HealthText(a))
	}
}

type containersByService []cloud66.Container

func (a containersByService) Len() int           { return len(a) }
func (a containersByService) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a containersByService) Less(i, j int) bool { return a[i].ServiceName < a[j].ServiceName }

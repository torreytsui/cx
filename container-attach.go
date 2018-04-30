package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/cloud66/cli"
)

func runContainerAttach(c *cli.Context) {
	if len(c.Args()) != 1 {
		cli.ShowSubcommandHelp(c)
		os.Exit(2)
	}

	fmt.Println("Attaching to container...")

	stack := mustStack(c)
	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()

	containerUid := c.Args()[0]
	container, err := client.GetContainer(stack.Uid, containerUid)
	must(err)

	if container == nil {
		printFatal("Container with Id '" + containerUid + "' not found")
	}

	var serverUID string
	if stack.Backend == "kubernetes" {
		servers, err := client.Servers(stack.Uid)
		must(err)

		kubernetesMasterServerUID := ""
		for _, server := range servers {
			if server.IsKubernetesMaster {
				kubernetesMasterServerUID = server.Uid
			}
		}

		if kubernetesMasterServerUID == "" {
			printFatal("Couldn't find a Kubernetes master server")
		}

		serverUID = kubernetesMasterServerUID
	} else {
		serverUID = container.ServerUid
	}

	server, err := client.GetServer(stack.Uid, serverUID, 0)
	must(err)

	var userCommand string
	if stack.Backend == "kubernetes" {
		namespace := stack.Namespaces[0]
		userCommand = fmt.Sprintf("kubectl --namespace=%s attach --stdin=false --tty=false %s", namespace, container.Uid)
	} else {
		userCommand = fmt.Sprintf("sudo docker attach --no-stdin=true --sig-proxy=false %s", container.Uid)
	}

	err = runServerCommand(*server, userCommand, false)
	if err != nil {
		printFatal(err.Error())
	}
}

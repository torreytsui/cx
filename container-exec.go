package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/cloud66/cli"
)

func runContainerExec(c *cli.Context) {
	if len(c.Args()) != 2 {
		cli.ShowSubcommandHelp(c)
		os.Exit(2)
	}

	fmt.Println("Running exec on container...")

	stack := mustStack(c)

	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()

	containerUid := c.Args()[0]
	command := c.Args()[1]
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

	cliFlags := c.String("cli-flags")

	var userCommand string
	if stack.Backend == "kubernetes" {
		if cliFlags == "" {
			cliFlags = "--stdin=true --tty=true"
		}

		namespace := stack.Namespaces[0]
		userCommand = fmt.Sprintf("kubectl --namespace=%s exec %s %s -- %s", namespace, cliFlags, container.Uid, command)
	} else {
		if cliFlags == "" {
			cliFlags = "--interactive=true --tty=true --detach=false"
		}

		userCommand = fmt.Sprintf("sudo docker exec %s %s %s", cliFlags, container.Uid, command)
	}

	err = SshToServerForCommand(*server, userCommand, false, true, "")
	if err != nil {
		printFatal(err.Error())
	}
}

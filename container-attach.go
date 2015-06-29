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

	server, err := client.GetServer(stack.Uid, container.ServerUid, 0)
	must(err)

	userCommand := fmt.Sprintf("sudo docker attach --no-stdin=true --sig-proxy=false %s", container.Uid)

	err = SshToServerForCommand(*server, userCommand, false)
	if err != nil {
		printFatal(err.Error())
	}
}

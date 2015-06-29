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

	server, err := client.GetServer(stack.Uid, container.ServerUid, 0)
	must(err)

	userCommand := fmt.Sprintf("sudo docker exec -it %s %s", container.Uid, command)

	err = SshToServerForCommand(*server, userCommand, "")
	if err != nil {
		printFatal(err.Error())
	}
}

package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/cloud66/cli"
)

func runContainerStart(c *cli.Context) {
	// get the server
	stack := mustStack(c)
	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()

	serverName := c.String("server")

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

	serviceName := c.String("service")
	commandToRun := c.String("command") // NOTE: not sure if it is working. Also have to pass the logs for the error down to the client if possible

	container, err := client.StartContainer(stack.Uid, server.Uid, serviceName, commandToRun)
	must(err)

	if container == nil {
		printFatal("Container could be started due to %s", err.Error())
	}

	fmt.Printf("Container %s started on %s\n", container.Uid, server.Name)

	attach := true
	noKill := false

	if c.IsSet("attach") {
		attach = c.Bool("attach")
	}
	if c.IsSet("no-kill") {
		noKill = c.Bool("no-kill")
	}

	if attach {
		fmt.Println("Attaching to the container...")
		userCommand := fmt.Sprintf("sudo docker attach --sig-proxy=false %s", container.Uid)

		err = SshToServerForCommand(*server, userCommand, "")
		if err != nil {
			printFatal(err.Error())
		}

		// NOTE: doesn't work since the signal is trapped by the SSH
		if !noKill {
			fmt.Println("Stopping container...")
			asyncId, err := startContainerStop(stack.Uid, container.Uid)
			if err != nil {
				printFatal(err.Error())
			}
			genericRes, err := endServerSet(*asyncId, stack.Uid)
			if err != nil {
				printFatal(err.Error())
			}
			printGenericResponse(*genericRes)
		}
	}

	return
}

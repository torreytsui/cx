package main

import (
	"fmt"
	"os"
	"os/signal"
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

		if !noKill {
			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt)
			go func() {
				for sig := range c {
					fmt.Printf("Stopping container with %s...\n", sig)
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
			}()
		}

		err = SshToServerForCommand(*server, userCommand, "")
		if err != nil {
			printFatal(err.Error())
		}
	}

	return
}

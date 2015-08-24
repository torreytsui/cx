package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/cloud66/cloud66"

	"github.com/cloud66/cli"
)

func runServiceInfo(c *cli.Context) {
	if len(c.Args()) != 1 {
		cli.ShowSubcommandHelp(c)
		os.Exit(2)
	}

	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()

	stack := mustStack(c)
	serviceName := c.Args()[0]

	flagServer := c.String("server")

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
		if !server.HasRole("docker") {
			printFatal("Server '" + flagServer + "' is not a docker server")
		}
		fmt.Printf("Server: %s\n", server.Name)
		serverUid = &server.Uid
	}

	service, err := client.GetService(stack.Uid, serviceName, serverUid, nil)
	must(err)

	printServiceInfoList(w, service)
	return
}

func printServiceInfoList(w io.Writer, service *cloud66.Service) {
	listRec(w, "NAME", "VALUE")
	listRec(w, "name", service.Name)
	listRec(w, "source type", service.SourceType)
	listRec(w, "git-ref", service.GitRef)
	listRec(w, "container count", strconv.Itoa(len(service.Containers)))
	listRec(w, "image name", service.ImageName)
	listRec(w, "image uid", service.ImageUid)
	listRec(w, "image tag", service.ImageTag)
	listRec(w, "command", service.Command)
	listRec(w, "build command", service.BuildCommand)
	listRec(w, "deploy command", service.DeployCommand)
}

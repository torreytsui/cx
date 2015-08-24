package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	// "sort"
	"text/tabwriter"

	"github.com/cloud66/cloud66"

	"github.com/cloud66/cli"
)

var cmdJobs = &Command{
	Name:       "jobs",
	Build:      buildJobs,
	Short:      "commands to work with jobs",
	NeedsStack: true,
}

func buildJobs() cli.Command {
	base := buildBasicCommand()
	base.Subcommands = []cli.Command{
		cli.Command{
			Name:   "list",
			Usage:  "lists all the jobs of a stack (or server)",
			Action: runJobs,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "server",
				},
				cli.StringFlag{
					Name: "service",
				},
			},
			Description: `List all the jobs of a stack or a server.

Examples:
$ cx jobs list -s mystack
$ cx jobs list -s mystack --server orca
$ cx jobs list -s mystack --server orca --service web
$ cx jobs list -s mystack --service web
`,
		},
		cli.Command{
			Name:   "run",
			Action: runJobRun,
			Usage:  "runs the given job once with given parameters",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "server",
				},
			},
			Description: `Runs the given job once with given parameters.
The list of available stack jobs can be obtained through the 'jobs list' command.

Examples:
$ cx jobs run -s mystack my_job
`,
		},
	}

	return base
}

func runJobs(c *cli.Context) {
	flagServer := c.String("server")
	flagServiceName := c.String("service")
	stack := mustStack(c)
	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()

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
			printFatal("Server '" + flagServer + "' can not host containers")
		}
		fmt.Printf("Server: %s\n", server.Name)
		serverUid = &server.Uid
	}

	var (
		jobs []cloud66.Job
		err  error
	)

	if flagServiceName == "" {
		jobs, err = client.GetJobs(stack.Uid, serverUid)
		must(err)
	}
	// else {
	// 	service, err := client.GetService(stack.Uid, flagServiceName, serverUid, nil)
	// 	must(err)
	// 	if service == nil {
	// 		printFatal("Service '" + flagServiceName + "' not found on specified stack")
	// 	} else {
	// 		services = make([]cloud66.Service, 1)
	// 		services[0] = *service
	// 	}
	// }
	printJobsList(w, jobs, flagServer)
}

func printJobsList(w io.Writer, jobs []cloud66.Job, flagServer string) {
	listRec(w,
		"JOB NAME",
		"TYPE",
		"CRON",
		"STATUS",
		"PARAMS",
	)

	// sort.Sort(ServiceByNameServer(services))
	for _, a := range jobs {
		listJob(w, a, flagServer)
	}
}

func listJob(w io.Writer, a cloud66.Job, flagServer string) {
	listRec(w,
		a.Name,
		a.Type,
		a.Cron,
		cloud66.JobStatus[a.Status],
		prettifyParams(a.Params),
	)
	// if len(a.Containers) != 0 {
	// 	for serverName, count := range a.ServerContainerCountMap() {
	// 		listRec(w,
	// 			a.Name,
	// 			serverName,
	// 			count,
	// 		)
	// 	}
	// } else if flagServer == "" {
	// 	listRec(w,
	// 		a.Name,
	// 		"n/a",
	// 		"0",
	// 	)
	// }

}

func prettifyParams(params map[string]string) string {
	res := ""
	for k, v := range params {
		k = strings.Replace(k, "_", " ", -1)
		res += fmt.Sprintf("%s: %s | ", k, v)
	}
	res = strings.TrimSuffix(res, "| ")

	return res
}

// type ServiceByNameServer []cloud66.Service

// func (a ServiceByNameServer) Len() int           { return len(a) }
// func (a ServiceByNameServer) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
// func (a ServiceByNameServer) Less(i, j int) bool { return a[i].Name < a[j].Name }

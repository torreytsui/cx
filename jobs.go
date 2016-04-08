package main

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/cloud66/cloud66"

	"github.com/cloud66/cli"
)

var cmdJobs = &Command{
	Name:       "jobs",
	Build:      buildJobs,
	Short:      "commands to work with jobs",
	NeedsStack: true,
	NeedsOrg:   false,
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
			Usage:  "runs the given job once with given parameters",
			Action: runJobRun,
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:  "arg",
					Value: &cli.StringSlice{},
				},
			},
			Description: `Runs the given job once with given parameters.
The list of available stack jobs can be obtained through the 'jobs list' command.

Examples:
$ cx jobs run -s mystack my_job
$ cx jobs run -s mystack --arg arg1 --arg arg2 my_job
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

	for _, a := range jobs {
		listJob(w, a, flagServer)
	}
}

func listJob(w io.Writer, a cloud66.Job, flagServer string) {
	switch a.(type) {
	case *cloud66.DockerHostTaskJob:
		DockerHostTaskJob(*a.(*cloud66.DockerHostTaskJob)).PrintList(w)
	case *cloud66.DockerServiceTaskJob:
		DockerServiceTaskJob(*a.(*cloud66.DockerServiceTaskJob)).PrintList(w)
	default:
		BasicJob(*a.(*cloud66.BasicJob)).PrintList(w)
	}
}

func (job BasicJob) PrintList(w io.Writer) {
	index := 0
	for k, v := range job.Params {
		index += 1
		if index == 1 {
			listRec(w,
				job.Name,
				job.Type,
				job.Cron,
				cloud66.JobStatus[job.Status],
				fmt.Sprintf("%s: %s", pascalCase(k, " "), v))
		} else {
			listRec(w, "", "", "", "", fmt.Sprintf("%s: %s", pascalCase(k, " "), v))
		}

	}
}

func (job DockerHostTaskJob) PrintList(w io.Writer) {
	listRec(w,
		job.Name,
		job.Type,
		job.Cron,
		cloud66.JobStatus[job.Status],
		fmt.Sprintf("Command: %s", job.Command),
	)
}

func (job DockerServiceTaskJob) PrintList(w io.Writer) {
	listRec(w,
		job.Name,
		job.Type,
		job.Cron,
		cloud66.JobStatus[job.Status],
		fmt.Sprintf("Task: %s", job.Task),
	)
	listRec(w, "", "", "", "", fmt.Sprintf("Service Name: %s", job.ServiceName))
	listRec(w, "", "", "", "", fmt.Sprintf("Private IP: %s", job.PrivateIp))
}

type BasicJob cloud66.BasicJob
type DockerHostTaskJob cloud66.DockerHostTaskJob
type DockerServiceTaskJob cloud66.DockerServiceTaskJob

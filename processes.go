package main

import "github.com/cloud66/cli"

var cmdProcesses = &Command{
	Name:       "processes",
	Build:      buildProcesses,
	Short:      "commands to work with processes",
	NeedsStack: true,
}

func buildProcesses() cli.Command {
	base := buildBasicCommand()
	base.Subcommands = []cli.Command{
		cli.Command{
			Name:   "list",
			Usage:  "lists all the processes of a stack (or server)",
			Action: runServices,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "server",
				},
				cli.StringFlag{
					Name: "service",
				},
			},
			Description: `List all the services and running containers of a stack or a server.

Examples:
$ cx services list -s mystack
$ cx services list -s mystack --server orca
$ cx services list -s mystack --server orca --service web
$ cx services list -s mystack --service web
`,
		},
		cli.Command{
			Name:   "scale",
			Action: runProcessScale,
			Usage:  "starts/stops processes from the given service",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "server",
				},
				cli.StringFlag{
					Name: "group",
				},
			},
			Description: `Starts <count> processes from the given process definition.
<count> can be an absolute value like "2" or a relative value like "+2" or "-3" etc.
If server is provided, will start <count> processes on that server.
If server is not provided, will start <count> processes on every server.

Examples:
$ cx processes scale -s mystack --server backend1 process_name +5
$ cx processes scale -s mystack --server backend2 process_name -5
$ cx processes scale -s mystack --server backend3 process_name 15
$ cx processes scale -s mystack process_name 1
`},
	}

	return base
}

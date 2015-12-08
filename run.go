package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/cloud66/cloud66"

	"github.com/cloud66/cli"
)

var cmdRun = &Command{
	Name:  "run",
	Build: buildBasicCommand,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "server",
			Usage: "server to run the command",
		},
		cli.StringFlag{
			Name:  "service",
			Usage: "name of the server to run the command in (docker stacks only)",
		},
	},
	Run:        runRun,
	NeedsStack: true,
	Short:      "executes a command directly on the server",
	Long: `This command will execute a command directly on the remote server.

For this purpose, this command will open the firewall for SSH from your IP address temporaritly (20 minutes), downloads the keys if you don't have them, starts a SSH session,
and executes the command specified.

If you have docker stack, you can provide and additional "service" argument, in which case the command will run in a new docker container based on the most recent service image.
Note that for docker stacks the command you provide is optional if your image already defines a CMD.

You need to have the correct access permissions to use this command.
You can use either the server name (ie lion) or the server IP (ie. 123.123.123.123) or the server role (ie. web)
with thie command.

If a role is specified the command will connect to the first server with that role.
Names are case insensitive and will work with the starting characters as well.

This command is only supported on Linux and OS X (for Windows you can run this in a virtual machine if necessary)

Examples:
$ cx run -s mystack --server lion 'ls -la'
$ cx run -s mystack --server 52.65.34.98 'ls -la'
$ cx run -s mystack --server web 'ls -la'
$ cx run -s mystack --server web --service api 'bundle exec rails c'
`,
}

func runRun(c *cli.Context) {
	if runtime.GOOS == "windows" {
		printFatal("Not supported on Windows")
		os.Exit(2)
	}

	stack := mustStack(c)
	if c.String("service") != "" && stack.Framework != "docker" {
		printFatal("The service option only applies to docker stacks")
		os.Exit(2)
	}

	serverName := c.String("server")

	if !c.IsSet("service") {
		if len(c.Args()) != 1 {
			cli.ShowCommandHelp(c, "run")
			os.Exit(2)
		}
	}

	userCommand := ""
	if len(c.Args()) == 1 {
		userCommand = c.Args()[0]
	}

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

	if c.String("service") != "" {
		// fetch service information for existing server/command
		service, err := client.GetService(stack.Uid, c.String("service"), &server.Uid, &userCommand)
		must(err)

		userCommand = service.WrapCommand
	}

	includeTty := c.String("service") != ""
	err = SshToServerForCommand(*server, userCommand, includeTty)
	if err != nil {
		printFatal(err.Error())
	}
}

func SshToServerForCommand(server cloud66.Server, userCommand string, includeTty bool) error {
	sshFile, err := prepareLocalSshKey(server)
	must(err)

	// open the firewall
	var timeToOpen = 2
	genericRes, err := client.LeaseSync(server.StackUid, nil, &timeToOpen, nil, &server.Uid)
	must(err)
	if genericRes.Status != true {
		printFatal("Unable to open server lease")
	}

	// add source
	userCommand = fmt.Sprintf("source /var/.cloud66_env &>/dev/null ; %s", userCommand)
	if includeTty {
		fmt.Println("Note: you may need to push <enter> to view output after the connection completes..")
		return startProgram("ssh", []string{
			server.UserName + "@" + server.Address,
			"-i", sshFile,
			"-o", "UserKnownHostsFile=/dev/null",
			"-o", "CheckHostIP=no",
			"-o", "StrictHostKeyChecking=no",
			"-o", "LogLevel=QUIET",
			"-o", "IdentitiesOnly=yes",
			"-A",
			"-p", "22",
			"-t",
			userCommand,
		})
	} else {
		return startProgram("ssh", []string{
			server.UserName + "@" + server.Address,
			"-i", sshFile,
			"-o", "UserKnownHostsFile=/dev/null",
			"-o", "CheckHostIP=no",
			"-o", "StrictHostKeyChecking=no",
			"-o", "LogLevel=QUIET",
			"-o", "IdentitiesOnly=yes",
			"-A",
			"-p", "22",
			userCommand,
		})
	}

}

package main

import (
	"os"
	"runtime"

	"github.com/cloud66/cloud66"
)

var cmdRun = &Command{
	Run:        runRun,
	Usage:      "run -s <stack> <server name>|<server ip>|<server role> '<command>'",
	NeedsStack: true,
	Category:   "stack",
	Short:      "executes a command directly on the server",
	Long: `This command will execute a command directly on the remote server.

For this purpose, this command will open the firewall for SSH from your IP address temporaritly (20 minutes), downloads the keys if you don't have them, starts a SSH session,
and executes the command specified.

You need to have the right access permissions to use this command.
You can use either the server name (ie lion) or the server IP (ie. 123.123.123.123) or the server role (ie. web)
with thie command.

If a role is specified the command will connect to the first server with that role.

Names are case insensitive and will work with the starting characters as well.

This command is only supported on Linux and OS X.

Examples:
$ cx run -s mystack lion 'ls -la'
$ cx run -s mystack 52.65.34.98 'ls -la'
$ cx run -s mystack web 'ls -la'
  `,
}

func runRun(cmd *Command, args []string) {
	if runtime.GOOS == "windows" {
		printFatal("Not supported on Windows")
		os.Exit(2)
	}

	stack := mustStack()

	// args start after stack name
	if len(args) != 2 {
		cmd.printUsage()
		os.Exit(2)
	}

	// get the server
	serverName := args[0]
	// get the user command
	userCommand := args[1]

	// fmt.Printf("User Command: %s\n", userCommand)

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

	err = sshToServerForCommand(*server, userCommand)
	if err != nil {
		printFatal(err.Error())
	}
}

func sshToServerForCommand(server cloud66.Server, userCommand string) error {
	sshFile, err := prepareLocalSshKey(server)
	must(err)

	// open the firewall
	var timeToOpen = 2
	genericRes, err := client.LeaseSync(server.StackUid, nil, &timeToOpen, nil, &server.Uid)
	must(err)
	if genericRes.Status != true {
		printFatal("Unable to open server lease")
	}

	return startProgram("ssh", []string{
		server.UserName + "@" + server.Address,
		"-i", sshFile,
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "CheckHostIP=no",
		"-o", "StrictHostKeyChecking=no",
		"-o", "LogLevel=QUIET",
		"-A",
		"-p", "22",
		userCommand,
	})
}

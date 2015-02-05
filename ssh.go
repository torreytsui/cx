package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/cloud66/cloud66"

	"github.com/cloud66/cli"
)

var cmdSsh = &Command{
	Name:       "ssh",
	Run:        runSsh,
	Build:      buildBasicCommand,
	NeedsStack: true,
	Short:      "starts a ssh shell into the server",
	Long: `This will open the firewall for SSH from your IP address temporaritly (20 minutes), downloads the keys if you don't have them
and starts a SSH session.

You need to have the right access permissions to use this command.
You can use either the server name (ie lion) or the server IP (ie. 123.123.123.123) or the server role (ie. web)
with thie command.

If a role is specified the command will connect to the first server with that role.

Names are case insensitive and will work with the starting characters as well.

This command is only supported on Linux and OS X.

Examples:
$ cx ssh -s mystack lion
$ cx ssh -s mystack 52.65.34.98
$ cx ssh -s mystack web
`,
}

func runSsh(c *cli.Context) {
	if runtime.GOOS == "windows" {
		printFatal("Not supported on Windows")
		os.Exit(2)
	}

	stack := mustStack(c)

	if len(c.Args()) != 1 {
		cli.ShowCommandHelp(c, "ssh")
		os.Exit(2)
	}

	// get the server
	serverName := c.Args()[0]

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

	fmt.Printf("Server: %s\n", server.Name)

	err = sshToServer(*server)
	if err != nil {
		printFatal(err.Error())
	}
}

func sshToServer(server cloud66.Server) error {
	sshFile, err := prepareLocalSshKey(server)
	must(err)

	// open the firewall
	timeToOpen := 20
	fmt.Printf("Opening access to %s...\n", server.Address)
	genericRes, err := client.LeaseSync(server.StackUid, nil, &timeToOpen, nil, &server.Uid)
	must(err)
	if genericRes.Status != true {
		printFatal("Unable to open server lease")
	}

	fmt.Printf("Connecting to %s (%s)...\n", server.Name, server.Address)
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
	})
}

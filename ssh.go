package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/cloud66/cx/cloud66"
)

var cmdSsh = &Command{
	Run:        runSsh,
	Usage:      "ssh <server name>|<server ip>|<server role>",
	NeedsStack: true,
	Category:   "stack",
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

func runSsh(cmd *Command, args []string) {
	if runtime.GOOS == "windows" {
		printFatal("Not supported on Windows")
		os.Exit(2)
	}

	stack := mustStack()

	if len(args) != 1 {
		cmd.printUsage()
		os.Exit(2)
	}

	// get the server
	serverName := args[0]

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
	sshFile := filepath.Join(homePath(), ".ssh", "cx_"+server.StackUid)

	// do we have the key?
	if b, _ := fileExists(sshFile); !b {
		// get the content and write the file
		fmt.Println("Fetching SSH key...")
		sshKey, err := client.ServerSshPrivateKey(server.Uid)

		if err != nil {
			return err
		}

		if err = writeSshFile(sshFile, sshKey); err != nil {
			return err
		}
	} else {
		if debugMode {
			fmt.Println("Found an existing SSH key for this server")
		}
	}

	// open the firewall
	fmt.Printf("Opening access to %s...\n", server.Address)
	_, err := client.Lease(server.StackUid, nil, nil, nil)
	must(err)
	if err != nil {
		return err
	}

	fmt.Printf("Connecting to %s (%s)...\n", server.Name, server.Address)
	return startProgram("ssh", []string{
		server.UserName + "@" + server.Address,
		"-i", sshFile,
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "CheckHostIP=no",
		"-o", "StrictHostKeyChecking=no",
		"-o", "LogLevel=QUIET",
		"-A",
		"-p", "22",
	})
}

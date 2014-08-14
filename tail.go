package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/cloud66/cloud66"
)

var cmdTail = &Command{
	Run:        runTail,
	Usage:      "tail <server name>|<server ip>|<server role> <log filename>",
	NeedsStack: true,
	Category:   "stack",
	Short:      "shows and tails the logfile specified on the given server",
	Long: `This will run a Linux tail command on the specified server and given logfile.
	Logs are read from stack's log folder (current/log) and should be the full logfile name
	including the extension.

	Server names and roles are case insensitive and will work with the starting characters as well.

	This command is only supported on Linux and OS X.

	Examples:

		$ cx tail -s mystack production.log
		$ cx ssh -s mystack 52.65.34.98 nginx_error.log
		$ cx ssh -s mystack web staging.log
	`,
}

func runTail(cmd *Command, args []string) {
	if runtime.GOOS == "windows" {
		printFatal("Not supported on Windows")
		os.Exit(2)
	}

	stack := mustStack()

	if len(args) != 2 {
		cmd.printUsage()
		os.Exit(2)
	}

	// get the server
	serverName := args[0]
	logName := args[1]

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

	err = tailLog(*stack, *server, logName)
	if err != nil {
		printFatal(err.Error())
	}
}

func tailLog(stack cloud66.Stack, server cloud66.Server, logName string) error {
	sshFile := filepath.Join(homePath(), ".ssh", "cx_"+server.StackUid)

	// do we have the key?
	if b, _ := fileExists(sshFile); !b {
		// get the content and write the file
		fmt.Println("Fetching SSH key...")
		sshKey, err := client.ServerSshPrivateKey(server.StackUid, server.Uid)

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
		fmt.Sprintf("tail -f '%s/web_head/current/log/%s'", stack.DeployDir, logName),
	})
}

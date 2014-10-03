package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/cloud66/cloud66"
)

var cmdDownload = &Command{
	Run:        runDownload,
	Usage:      "download -s <stack> <server name>|<server ip>|<server role> /path/to/source/file (optional: /path/to/target/directory)",
	NeedsStack: true,
	Category:   "stack",
	Short:      "copies a file from the remote server to your local computer",
	Long: `This command will copy a file from the remote server to your local computer.

This command will download the files to the home directory by default.
To copy the file to a specific directory in your local computer,
specify the target directory location in the command line.

This will open the firewall for SSH from your IP address temporaritly (20 minutes), downloads the keys if you don't have them
and starts a SSH session.

You need to have the right access permissions to use this command.
You can use either the server name (ie lion) or the server IP (ie. 123.123.123.123) or the server role (ie. web)
with thie command.

If a role is specified the command will connect to the first server with that role.

Names are case insensitive and will work with the starting characters as well.

This command is only supported on Linux and OS X.

Examples:
$ cx download -s mystack lion /path/to/source/file /path/to/target/directory
$ cx download -s mystack 52.65.34.98 /path/to/file
$ cx download -s mystack 52.65.34.98 /path/to/source/file /path/to/target/directory
`,
}

func runDownload(cmd *Command, args []string) {
	if runtime.GOOS == "windows" {
		printFatal("Not supported on Windows")
		os.Exit(2)
	}

	stack := mustStack()

	// args start after stack name
	// and check if user specified target directory
	var targetDirectory string = ""

	if len(args) < 2 {
		cmd.printUsage()
		os.Exit(2)
	} else if len(args) == 3 {
		targetDirectory = args[2]
	}

	// get the server
	serverName := args[0]
	// get the file path
	filePath := args[1]

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

	if targetDirectory == "" {
		err = sshToServerToDownload(*server, filePath)
	} else {
		err = sshToServerToDownload(*server, filePath, targetDirectory)
	}

	if err != nil {
		printFatal(err.Error())
	}
}

func sshToServerToDownload(server cloud66.Server, filePath string, targetDirectory ...string) error {
	// default target directory
	var defaultDir string = "."
	var targetDir string = defaultDir

	// if target directory specified
	if len(targetDirectory) > 0 {
		targetDir = targetDirectory[0]
	}

	sshFile, err := prepareLocalSshKey(server)
	must(err)

	// open the firewall
	var timeToOpen = 2
	fmt.Printf("Opening access to %s...\n", server.Address)
	genericRes, err := client.LeaseSync(server.StackUid, nil, &timeToOpen, nil)
	must(err)
	if genericRes.Status != true {
		printFatal("Unable to open server lease")
	}

	fmt.Printf("Connecting to %s (%s)...\n", server.Name, server.Address)

	return startProgram("scp", []string{
		"-i", sshFile,
		"-r",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "CheckHostIP=no",
		"-o", "StrictHostKeyChecking=no",
		"-o", "LogLevel=QUIET",
		"-P", "22",
		server.UserName + "@" + server.Address + ":" + filePath,
		targetDir,
	})
}

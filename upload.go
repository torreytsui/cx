package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	
	"github.com/cloud66/cx/cloud66"
)

var cmdUpload = &Command{
	Run:        runUpload,
	Usage:      "upload -s <stack> <server> /path/to/source/file (optional: /path/to/target/directory)",
	NeedsStack: true,
	Category:   "stack",
	Short:      "copies a file from your local computer to the remote server",
	Long: `This command will copy a file from your local computer to the remote server. 

  This command will upload the files to the '/tmp' directory by default. 
  To copy the file to a specific directory in the remote server,
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

    $ cx upload -s mystack lion /path/to/source/file
    $ cx upload -s mystack 52.65.34.98 /path/to/source/file
    $ cx upload -s mystack 52.65.34.98 /path/to/source/file /path/to/target/directory
  `,
}

func runUpload(cmd *Command, args []string) {
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
		err = sshToServerToUpload(*server, filePath)
	} else {
		err = sshToServerToUpload(*server, filePath, targetDirectory)
	}
	
	if err != nil {
		printFatal(err.Error())
	}
}

func sshToServerToUpload(server cloud66.Server, filePath string, targetDirectory ...string) error {
	sshFile := filepath.Join(homePath(), ".ssh", "cx_"+server.StackUid)

	// default target directory
	var defaultDir string = "/tmp"
	var targetDir string = defaultDir
	
	// if target directory specified
	if len(targetDirectory) > 0 {
		targetDir = targetDirectory[0]
	}
	
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
	
	return startProgram("scp", []string{
		"-i", sshFile,
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "CheckHostIP=no",
		"-o", "StrictHostKeyChecking=no",
		"-o", "LogLevel=QUIET",
		"-P", "22",
		filePath,
		server.UserName + "@" + server.Address + ":" + targetDir,
	})
}
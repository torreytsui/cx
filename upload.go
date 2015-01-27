package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/cloud66/cloud66"

	"github.com/codegangsta/cli"
)

var cmdUpload = &Command{
	Run:   runUpload,
	Build: buildBasicCommand,
	Name:  "upload",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "server",
			Usage: "server to upload to",
		},
	},
	NeedsStack: true,
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
$ cx upload -s mystack --server lion /path/to/source/file
$ cx upload -s mystack --server lion /path/to/source/file /path/to/target/directory
$ cx upload -s mystack --server 52.65.34.98 /path/to/source/file
$ cx upload -s mystack --server web /path/to/source/file /path/to/target/directory
`,
}

func runUpload(c *cli.Context) {
	if runtime.GOOS == "windows" {
		printFatal("Not supported on Windows")
		os.Exit(2)
	}

	stack := mustStack(c)

	// args start after stack name
	// and check if user specified target directory
	var targetDirectory string = ""

	if len(c.Args()) < 1 {
		cli.ShowCommandHelp(c, "upload")
		os.Exit(2)
	} else if len(c.Args()) == 2 {
		targetDirectory = c.Args()[1]
	}

	// get the server
	serverName := c.String("server")
	// get the file path
	filePath := c.Args()[0]

	servers, err := client.Servers(stack.Uid)
	if err != nil {
		printFatal(err.Error())
	}

	server, err := findServer(servers, serverName)
	if err != nil {
		printError("server not found, please ensure correct server is specified in command.")
		os.Exit(2)
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
	// default target directory
	var defaultDir string = "/tmp"
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
	genericRes, err := client.LeaseSync(server.StackUid, nil, &timeToOpen, nil, &server.Uid)
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
		"-o", "IdentitiesOnly=yes",
		"-P", "22",
		filePath,
		server.UserName + "@" + server.Address + ":" + targetDir,
	})
}

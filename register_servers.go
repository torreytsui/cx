package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/cloud66/cli"
)

var cmdRegisterServer = &Command{
	Name:  "register-server",
	Build: buildBasicCommand,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "server",
			Usage: "server to register",
		},
		cli.StringFlag{
			Name:  "key",
			Usage: "Private SSH key to connect to the servers",
		},
		cli.StringFlag{
			Name:  "file",
			Usage: "file containing server addresses (one server per line)",
		},
		cli.StringFlag{
			Name:  "user",
			Usage: "username for connecting to the server",
		},
		cli.StringFlag{
			Name:  "force-local-ip",
			Usage: "force-local-ip for using the local ip address of the registered server",
		},
	},
	Run:        runRegisterServer,
	NeedsStack: false,
	NeedsOrg:   true,
	Short:      "registers given server with the account",
	Long: `This command is a shortcut to run the server registration script on a group of servers


Example:
$ cx register-server --org team --user root --server 149.56.134.22 --key ~/.ssh/private_key
$ cx register-server --org team --user ubuntu --file servers.txt --key ~/.ssh/private_key
$ cx register-server --org team --user ubuntu --file servers.txt --key ~/.ssh/private_key --force-local-ip true
`,
}

func runRegisterServer(c *cli.Context) {
	if runtime.GOOS == "windows" {
		printFatal("Not supported on Windows")
		os.Exit(2)
	}


	org := mustOrg(c)
	info, err := client.AccountInfo(org.Id, false)
	if err != nil {
		printFatal(err.Error())
	}

	regScript := info.ServerRegistration
	if regScript == "" {
		printFatal("Unable to fetch registration script for this account")
	}

	if c.String("file") == "" && c.String("server") == "" {
		printFatal("Either --file or --server should be available")
	}

	if c.String("file") != "" && c.String("server") != "" {
		printFatal("Cannot use both --file and --server at the same time")
	}

	if c.String("user") == "" {
		printFatal("No --user specified")
	}

	use_local_ip := c.String("force-local-ip") == "true"


	if c.String("server") != "" {
		if err = registerServer(c.String("server"), regScript, c.String("key"), c.String("user"), use_local_ip); err != nil {
			printFatal(err.Error())
		}
	} else if c.String("file") != "" {
		file, err := os.Open(c.String("file"))
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			s := scanner.Text()
			if err = registerServer(s, regScript, c.String("key"), c.String("user"), use_local_ip); err != nil {
				printError("Failed to register %s due to %s", s, err.Error())
			}
		}
		if err := scanner.Err(); err != nil {
			printFatal(err.Error())
		}
	}
	fmt.Printf("Register server(s) done.\n")
	

}

func registerServer(server string, script string, keyFile string, user string, use_local_ip bool) error {
	extra_header := ""
	if use_local_ip {
		extra_header = "--header \"X-Force-Local-IP:true\""
	}
	toRun := fmt.Sprintf("'curl %s -s %s| bash -s'", extra_header, script)
	if keyFile == "" {
		return startProgram("ssh", []string{
			user + "@" + server,
			"-o", "UserKnownHostsFile=/dev/null",
			"-o", "CheckHostIP=no",
			"-o", "StrictHostKeyChecking=no",
			"-o", "LogLevel=QUIET",
			"-o", "IdentitiesOnly=yes",
			"-A",
			"-t",
			"-p", "22",
			"sudo su - -c",
			toRun,
		})
	} else {
		return startProgram("ssh", []string{
			user + "@" + server,
			"-i", keyFile,
			"-o", "UserKnownHostsFile=/dev/null",
			"-o", "CheckHostIP=no",
			"-o", "StrictHostKeyChecking=no",
			"-o", "LogLevel=QUIET",
			"-o", "IdentitiesOnly=yes",
			"-A",
			"-t",
			"-p", "22",
			"sudo su - -c",
			toRun,
		})
	}
}

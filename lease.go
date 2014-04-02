package main

import (
	"fmt"
)

var cmdLease = &Command{
	Run:        runLease,
	Usage:      "lease [-f <from IP>] [-t <time to open>] [-p <port>]",
	NeedsStack: true,
	Category:   "stack",
	Short:      "leases firewall access to the given server on the stack",
	Long: `This will poke a hole in the firewall of the given server for a limited time.
  'Time to open' is in minutes (ie. 60 for 1 hour)
  By default the firewall is closed after 20 minutes. The maximum time to open is 240 minutes.

  If no 'from IP' is specified, the caller's IP address (your IP address) is used.
  If no 'port' is used, the default is 22 (SSH).

Examples:

    $ cx lease -s mystack
    $ cx lease -s mystack -t 120 -p 3306
    $ cx lease -s mystack -p 3306 -f 52.65.34.98
`,
}

var (
	flagTimeToOpen int
	flagPort       int
	flagIp         string
)

func init() {
	cmdLease.Flag.IntVar(&flagTimeToOpen, "t", 20, "time to open")
	cmdLease.Flag.IntVar(&flagPort, "p", 22, "port")
	cmdLease.Flag.StringVar(&flagIp, "f", "", "from IP")
}

func runLease(cmd *Command, args []string) {
	stack := mustStack()

	result, err := client.Lease(stack.Uid, &flagIp, &flagTimeToOpen, &flagPort)
	must(err)

	if err != nil {
		printFatal(err.Error())
	} else {
		fmt.Println(result.Message)
	}

}

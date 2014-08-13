package main

import "time"

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
	cmdLease.Flag.StringVar(&flagIp, "f", "AUTO", "from IP")
}

func runLease(cmd *Command, args []string) {
	stack := mustStack()

	async_result, err := client.Lease(stack.Uid, &flagIp, &flagTimeToOpen, &flagPort)
	var async_error = client.WaitForAsyncActionComplete(stack.Uid, async_result, err, 2*time.Second, 2*time.Minute, true)
	if async_error != nil {
		printFatal(async_error.Error())
	}
}

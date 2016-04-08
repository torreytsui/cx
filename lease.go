package main

import (
	"fmt"

	"github.com/cloud66/cli"
)

var cmdLease = &Command{
	Run:   runLease,
	Name:  "lease",
	Build: buildBasicCommand,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "from,f",
			Usage: "IP address for the source of traffic. Uses your current IP if not provided",
			Value: "AUTO",
		},
		cli.IntFlag{
			Name:  "tto,t",
			Usage: "Time to keep the lease open",
			Value: 20,
		},
		cli.IntFlag{
			Name:  "port,p",
			Usage: "Port to open",
			Value: 22,
		},
	},
	NeedsStack: true,
	NeedsOrg:   false,
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

func runLease(c *cli.Context) {
	stack := mustStack(c)

	from := c.String("from")
	tto := c.Int("tto")
	port := c.Int("port")

	fmt.Printf("Attempting to lease from %s to port %d for %d minutes...\n", from, port, tto)
	genericRes, err := client.LeaseSync(stack.Uid, &from, &tto, &port, nil)
	if err != nil {
		printFatal(err.Error())
	}
	printGenericResponse(*genericRes)
}

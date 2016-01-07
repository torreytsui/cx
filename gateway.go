package main

import (
	//	"fmt"
	"github.com/cloud66/cli"
	//	"io/ioutil"
	"os"
	"text/tabwriter"
)

var cmdGateway = &Command{
	NeedsStack: false,
	Build:      buildGateways,
	Name:       "gateways",
	Short:      "commands to work with gateways",
}

func buildGateways() cli.Command {
	base := buildBasicCommand()
	base.Subcommands = []cli.Command{
		cli.Command{
			Name:   "list",
			Action: runListGateways,
			Usage:  "gateways list",
			Description: `lists gateways of an account
Examples:
$ cx gateways list
`,
		},
		cli.Command{
			Name:   "add",
			Action: runAddGateway,
			Usage:  "gateways add <gateway name> --address <gateway address> --username <gateway username>",
			Description: `Add gateway to an account.
				Examples:
				$ cx gateways add aws_bastion --address 192.168.100.100  --username ec2-user
				`,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "address",
				},
				cli.StringFlag{
					Name: "username",
				},
			},
		},
		/*				cli.Command{
									Name:   "open",
									Action: runOpenGateway,
									Usage:  "gateways open <gateway name> --key <path/to/gateway/key/file>  --ttl <time to live>",
									Description: `Make the gateway available to use with cloud66 for ttl amount of time.

						Example:
						$ cx gateways open aws_bastion --key /tmp/gateway.pem --ttl 1800`,
									Flags: []cli.Flag{
										cli.StringFlag{
											Name: "key",
										},
										cli.IntFlag{
											Name: "ttl",
										},
									},
								},
								cli.Command{
									Name:   "remove",
									Action: runRemoveGateway,
									Usage:  "gateways remove <gateway name>",
									Description: `Remove gateway from an account.

						Example:
						$ cx gateways remove aws_bastion`,
								},
		*/}

	return base
}

func runListGateways(c *cli.Context) {
	currentAccountId := findAccountId()
	if currentAccountId == 0 {
		printFatal("Can not find current account")
		os.Exit(2)
	}

	gateways, err := client.ListGateways(currentAccountId)
	if err != nil {
		printFatal("Error list gateways: " + err.Error())
		os.Exit(2)
	}

	if gateways != nil {
		w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
		defer w.Flush()

		for _, g := range gateways {
			listRec(w,
				g.Id,
				g.Name,
				g.Username,
				g.Address,
				g.Ttl,
				g.Content,
				g.CreatedAt,
			)
		}

	}

}

func runAddGateway(c *cli.Context) {
	if len(c.Args()) == 0 {
		printFatal("You should specify a name for gateway")
		os.Exit(2)
	}
	name := c.Args()[0]

	flagAddress := ""

	if c.IsSet("address") {
		flagAddress = c.String("address")
	} else {
		printFatal("You should specify an address for gateway")
		os.Exit(2)
	}

	flagUsername := ""

	if c.IsSet("username") {
		flagUsername = c.String("username")
	} else {
		printFatal("You should specify a username for gateway")
		os.Exit(2)
	}

	currentAccountId := findAccountId()
	if currentAccountId == 0 {
		printFatal("Can not find current account")
		os.Exit(2)
	}

	err := client.AddGateway(currentAccountId, name, flagAddress, flagUsername)

	if err != nil {
		printFatal("Error adding gateway : " + err.Error())
		os.Exit(2)
	}
}

/*
func runOpenGateway(c *cli.Context) {
	currentAccountId := findAccountId()
	if currentAccountId == 0 {
		printFatal("Can not find current account")
		os.Exit(2)
	}

	err := client.RemoveGatewayKey(currentAccountId)

	if err != nil {
		printFatal("Error remove gateway key : " + err.Error())
		os.Exit(2)
	}
}

func runRemoveGateway(c *cli.Context) {
	currentAccountId := findAccountId()
	if currentAccountId == 0 {
		printFatal("Can not find current account")
		os.Exit(2)
	}

	err := client.RemoveGatewayKey(currentAccountId)

	if err != nil {
		printFatal("Error remove gateway key : " + err.Error())
		os.Exit(2)
	}
}
*/
func findAccountId() int {
	accountInfos, err := client.AccountInfos()
	if err != nil {
		printFatal("Error requesting account info : " + err.Error())
		os.Exit(2)
	}

	if len(accountInfos) < 1 {
		printFatal("User associated with this request returning zero references")
		os.Exit(2)
	}

	var currentAccountId int
	currentAccountId = 0

	for _, accountInfo := range accountInfos {

		if accountInfo.CurrentAccount {
			currentAccountId = accountInfo.Id
		}
	}

	return currentAccountId
}

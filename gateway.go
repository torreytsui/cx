package main

import (
	"fmt"
	"github.com/cloud66/cli"
	"io/ioutil"
	"os"
	"strings"
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
			Usage:  "gateways add --name <gateway name> --address <gateway address> --username <gateway username>",
			Description: `Add gateway to an account.
Examples:
$ cx gateways add --name aws_bastion --address 192.168.100.100  --username ec2-user
				`,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "name",
				},
				cli.StringFlag{
					Name: "address",
				},
				cli.StringFlag{
					Name: "username",
				},
			},
		},
		cli.Command{
			Name:   "open",
			Action: runOpenGateway,
			Usage:  "gateways open --name <gateway name> --key <path/to/gateway/key/file>  --ttl <time to live>",
			Description: `Make the gateway available to use with cloud66 for ttl amount of time.
Example:
$ cx gateways open aws_bastion --key /tmp/gateway.pem --ttl 1800`,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "name",
				},
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
			Usage:  "gateways remove --name <gateway name>",
			Description: `Remove gateway from an account.
Example:
$ cx gateways remove aws_bastion`,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "name",
				},
			},
		},
	}

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
		printFatal("Error listing gateways: " + err.Error())
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
	gatewayName := ""
	if c.IsSet("name") {
		gatewayName = c.String("name")
	} else {
		printFatal("You should specify a name for gateway\ngateways add --name <gateway name> --address <gateway address> --username <gateway username>")
		os.Exit(2)
	}

	flagAddress := ""

	if c.IsSet("address") {
		flagAddress = c.String("address")
	} else {
		printFatal("You should specify an address for gateway\ngateways add --name <gateway name> --address <gateway address> --username <gateway username>")
		os.Exit(2)
	}

	flagUsername := ""

	if c.IsSet("username") {
		flagUsername = c.String("username")
	} else {
		printFatal("You should specify a username for gateway\ngateways add --name <gateway name> --address <gateway address> --username <gateway username>")
		os.Exit(2)
	}

	currentAccountId := findAccountId()
	if currentAccountId == 0 {
		printFatal("Can not find current account")
		os.Exit(2)
	}

	err := client.AddGateway(currentAccountId, gatewayName, flagAddress, flagUsername)

	if err != nil {
		printFatal("Error adding gateway : " + err.Error())
		os.Exit(2)
	}
}

func runOpenGateway(c *cli.Context) {
	gatewayName := ""
	if c.IsSet("name") {
		gatewayName = c.String("name")
	} else {
		printFatal("You should specify the name of gateway to open\ngateways open --name <gateway name> --key <path/to/gateway/key/file>  --ttl <time to live>")
		os.Exit(2)
	}

	currentAccountId := findAccountId()
	if currentAccountId == 0 {
		printFatal("Can not find current account")
		os.Exit(2)
	}

	gatewayId := findGatwayId(currentAccountId, gatewayName)

	flagTtl := 0

	if c.IsSet("ttl") {
		flagTtl = c.Int("ttl")
	} else {
		printFatal("You should specify a ttl for gateway to be open\ngateways open --name <gateway name> --key <path/to/gateway/key/file>  --ttl <time to live>")
		os.Exit(2)
	}

	flagKeyFile := ""

	if c.IsSet("key") {
		flagKeyFile = c.String("key")
	} else {
		printFatal("You should specify a key file path\ngateways open --name <gateway name> --key <path/to/gateway/key/file>  --ttl <time to live>")
		os.Exit(2)
	}

	keyfilePath := expandPath(flagKeyFile)
	keyContent, err := ioutil.ReadFile(keyfilePath)
	if err != nil {
		printFatal("Can not read from %s : "+err.Error(), keyfilePath)
		os.Exit(2)
	}

	err = client.UpdateGateway(currentAccountId, gatewayId, string(keyContent), flagTtl)

	if err != nil {
		printFatal("Error opening gateway : " + err.Error())
		os.Exit(2)
	}

}

func runRemoveGateway(c *cli.Context) {
	gatewayName := ""
	if c.IsSet("name") {
		gatewayName = c.String("name")
	} else {
		printFatal("You should specify the name of gateway to delete\ngateways remove --name <gateway name>")
		os.Exit(2)
	}

	currentAccountId := findAccountId()
	if currentAccountId == 0 {
		printFatal("Can not find current account")
		os.Exit(2)
	}

	gatewayId := findGatwayId(currentAccountId, gatewayName)

	err := client.RemoveGateway(currentAccountId, gatewayId)

	if err != nil {
		printFatal("Error remove gateway : " + err.Error())
		os.Exit(2)
	}
}

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
			break
		}
	}

	return currentAccountId
}

func findGatwayId(accountId int, gatewayName string) int {
	gateways, err := client.ListGateways(accountId)
	if err != nil {
		printFatal("Error query list of gateways: " + err.Error())
		os.Exit(2)
	}

	result := -1
	for _, g := range gateways {
		if strings.Compare(g.Name, gatewayName) == 0 {
			result = g.Id
			break
		}
	}

	if result == -1 {
		printFatal(fmt.Sprintf("Can not find gateway(%s)", gatewayName))
		os.Exit(2)
	}
	return result
}

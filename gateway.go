package main

import (
	"fmt"
	"github.com/cloud66/cli"
	"github.com/cloud66/cloud66"
	"io/ioutil"
	"os"
	"strings"
	"text/tabwriter"
	"time"
)

var cmdGateway = &Command{
	NeedsStack: false,
	NeedsOrg:   false,
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
$ cx gateways list --verbose
`, Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "verbose",
					Usage: "Show more information about each gateway",
				},
			},
		},
		cli.Command{
			Name:   "add",
			Action: runAddGateway,
			Usage:  "gateways add --name <gateway name> --address <gateway address> --username <gateway username>  --private-ip <private ip of gateway>",
			Description: `Add gateway to an account.
Examples:
$ cx gateways add --name aws_bastion --address 192.168.100.100  --username ec2-user  --private-ip 192.168.1.1
				`,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "name,n",
				},
				cli.StringFlag{
					Name: "address",
				},
				cli.StringFlag{
					Name: "username",
				},
				cli.StringFlag{
					Name: "private-ip",
				},
			},
		},
		cli.Command{
			Name:   "open",
			Action: runOpenGateway,
			Usage:  "gateways open --name <gateway name> --key <path/to/gateway/key/file>  --ttl <time to live  1h, 30m, 30s, ...>",
			Description: `Make the gateway available to use with cloud66 for ttl amount of time.
Example:
$ cx gateways open aws_bastion --key /tmp/gateway.pem --ttl 45m`,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "name,n",
				},
				cli.StringFlag{
					Name: "key",
				},
				cli.StringFlag{
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
					Name: "name,n",
				},
			},
		},
		cli.Command{
			Name:   "close",
			Action: runCloseGateway,
			Usage:  "gateways close --name <gateway name>",
			Description: `Close a gateway.
Example:
$ cx gateways close aws_bastion`,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "name,n",
				},
			},
		},
	}

	return base
}

func runListGateways(c *cli.Context) {
	org := mustOrg(c)
	flagVerbose := c.Bool("verbose")

	result := []cloud66.Gateway{}

	gateways, err := client.ListGateways(org.Id)
	if err != nil {
		printFatal("Error listing gateways: " + err.Error())
		os.Exit(2)
	}
	for _, g := range gateways {
		result = append(result, g)
	}

	if result != nil {
		w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
		defer w.Flush()

		if len(result) == 0 {
			fmt.Println("No gateway defined.")
		} else {
			if flagVerbose {
				listRec(w, "NAME", "USERNAME", "ADDRESS", "PRIVATE IP", "OPEN UNTIL", "STATE", "CREATED AT", "CREATED BY", "UPDATED AT", "UPDATED BY")
			} else {
				listRec(w, "NAME", "USERNAME", "ADDRESS", "PRIVATE IP", "OPEN UNTIL", "STATE")
			}

			for _, g := range result {
				state := "N/A"
				if (len(g.Content) > 0) && (strings.Compare(g.Content, "N/A") != 0) {
					state = "open"
				} else {
					state = "close"
				}

				ttl_string := "N/A"
				t, err := time.Parse("2006-01-02T15:04:05Z", g.Ttl)
				if err == nil {
					ttl_string = prettyTime{t}.String()
				}

				if flagVerbose {
					listRec(w,
						g.Name,
						g.Username,
						g.Address,
						g.PrivateIp,
						ttl_string,
						state,
						prettyTime{g.CreatedAt},
						g.CreatedBy,
						prettyTime{g.UpdatedAt},
						g.UpdatedBy,
					)
				} else {
					listRec(w,
						g.Name,
						g.Username,
						g.Address,
						g.PrivateIp,
						ttl_string,
						state,
					)
				}
			}

		}
	}

}

func runAddGateway(c *cli.Context) {
	gatewayName := ""
	if c.IsSet("name") {
		gatewayName = c.String("name")
	} else {
		printFatal("You should specify a name for gateway\ngateways add --name <gateway name> --address <gateway address> --username <gateway username>  --private-ip <private ip of gateway>")
		os.Exit(2)
	}

	flagAddress := ""

	if c.IsSet("address") {
		flagAddress = c.String("address")
	} else {
		printFatal("You should specify an address for gateway\ngateways add --name <gateway name> --address <gateway address> --username <gateway username>  --private-ip <private ip of gateway>")
		os.Exit(2)
	}

	flagUsername := ""

	if c.IsSet("username") {
		flagUsername = c.String("username")
	} else {
		printFatal("You should specify a username for gateway\ngateways add --name <gateway name> --address <gateway address> --username <gateway username>   --private-ip <private ip of gateway>")
		os.Exit(2)
	}

	flagPrivateIp := ""

	if c.IsSet("private-ip") {
		flagPrivateIp = c.String("private-ip")
	} else {
		printFatal("You should specify private ip of gateway\ngateways add --name <gateway name> --address <gateway address> --username <gateway username>  --private-ip <private ip of gateway>")
		os.Exit(2)
	}

	currentAccountId := findAccountId(c)
	if currentAccountId == 0 {
		printFatal("Can not find current account")
		os.Exit(2)
	}

	err := client.AddGateway(currentAccountId, gatewayName, flagAddress, flagUsername, flagPrivateIp)

	if err != nil {
		printFatal("Error adding gateway : " + err.Error())
		os.Exit(2)
	}
	fmt.Println("Gateway added successfully!")
}

func runOpenGateway(c *cli.Context) {
	gatewayName := ""
	if c.IsSet("name") {
		gatewayName = c.String("name")
	} else {
		printFatal("You should specify the name of gateway to open\ngateways open --name <gateway name> --key <path/to/gateway/key/file>  --ttl <time to live  1h, 30m, 30s, ...>")
		os.Exit(2)
	}

	accountId, gatewayId, state := findGatwayInfo(c, gatewayName)

	if strings.Compare(state, "open") == 0 {
		fmt.Println("Gateway is already open.")
	} else {
		ttlValue := 0

		if c.IsSet("ttl") {
			flagTtl := c.String("ttl")

			d, e := time.ParseDuration(flagTtl)
			if e != nil {
				printFatal("Wrong TTL format : " + e.Error())
				os.Exit(2)
			}
			ttlValue = int(d.Seconds())
		} else {
			printFatal("You should specify a ttl for gateway to be open\ngateways open --name <gateway name> --key <path/to/gateway/key/file>  --ttl <time to live  1h, 30m, 30s, ...>")
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

		err = client.UpdateGateway(accountId, gatewayId, string(keyContent), ttlValue)

		if err != nil {
			printFatal("Error opening gateway : " + err.Error())
			os.Exit(2)
		}
		fmt.Println("Gateway opened successfully!")
	}

}

func runRemoveGateway(c *cli.Context) {
	gatewayName := ""
	if c.IsSet("name") {
		gatewayName = c.String("name")
	} else {
		printFatal("You should specify the name of gateway to remove\ngateways remove --name <gateway name>")
		os.Exit(2)
	}

	accountId, gatewayId, state := findGatwayInfo(c, gatewayName)

	if strings.Compare(state, "open") == 0 {
		printFatal("Can not remove an open gateway, please close it first.")
	} else {
		err := client.RemoveGateway(accountId, gatewayId)

		if err != nil {
			printFatal("Error remove gateway : " + err.Error())
			os.Exit(2)
		}
		fmt.Println("Gateway removed successfully!")
	}
}

func runCloseGateway(c *cli.Context) {
	gatewayName := ""
	if c.IsSet("name") {
		gatewayName = c.String("name")
	} else {
		printFatal("You should specify the name of gateway to close\ngateways close --name <gateway name>")
		os.Exit(2)
	}

	accountId, gatewayId, state := findGatwayInfo(c, gatewayName)

	if strings.Compare(state, "close") == 0 {
		fmt.Println("Gateway is already closed.")
	} else {
		flagTtl := 1
		keyContent := ""

		err := client.UpdateGateway(accountId, gatewayId, string(keyContent), flagTtl)

		if err != nil {
			printFatal("Error close gateway : " + err.Error())
			os.Exit(2)
		}
		fmt.Println("Gateway closed successfully!")
	}
}

func findAccountId(c *cli.Context) int {
	account := mustOrg(c)

	return account.Id
}

func findGatwayInfo(c *cli.Context, gatewayName string) (int, int, string) {
	org := mustOrg(c)

	resultGatewayId := -1
	resultAccountId := -1
	resultState := "N/A"

	gateways, err := client.ListGateways(org.Id)
	if err == nil {
		for _, g := range gateways {
			if strings.Compare(g.Name, gatewayName) == 0 {
				resultGatewayId = g.Id
				resultAccountId = org.Id
				if (len(g.Content) > 0) && (strings.Compare(g.Content, "N/A") != 0) {
					resultState = "open"
				} else {
					resultState = "close"
				}
				break
			}
		}
	}

	if resultGatewayId == -1 {
		printFatal(fmt.Sprintf("Can not find gateway \"%s\"", gatewayName))
		os.Exit(2)
	}

	return resultGatewayId, resultAccountId, resultState
}

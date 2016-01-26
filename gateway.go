package main

import (
	"fmt"
	"github.com/cloud66/cli"
	"github.com/cloud66/cloud66"
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
			Usage:  "gateways add --name <gateway name> --address <gateway address> --username <gateway username>  --private-ip <private ip of gateway>",
			Description: `Add gateway to an account.
Examples:
$ cx gateways add --name aws_bastion --address 192.168.100.100  --username ec2-user  --private-ip 192.168.1.1
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
				cli.StringFlag{
					Name: "private-ip",
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
		cli.Command{
			Name:   "close",
			Action: runCloseGateway,
			Usage:  "gateways close --name <gateway name>",
			Description: `Close a gateway.
Example:
$ cx gateways close aws_bastion`,
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
	accountInfos, err := client.AccountInfos()
	if err != nil {
		printFatal("Can not retrive account info.")
		os.Exit(2)
	}
    
	result := []cloud66.Gateway{}
	
	for _, accountInfo := range accountInfos {		
		if accountInfo.CurrentAccount {
			gateways, err := client.ListGateways(accountInfo.Id)
			if err != nil {
				printFatal("Error listing gateways: " + err.Error())
				os.Exit(2)
			}
			for _, g := range gateways {
				result = append(result, g)
			}
		} else {
			gateways, err := client.ListGateways(accountInfo.Id)
			if err == nil {
				for _, g := range gateways {
					result = append(result, g)
				}
			}
		}		
	}


	if result != nil {
		w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
		defer w.Flush()

		if len(result) == 0 {
			fmt.Println("No gateway defined.")
		} else {
			listRec(w,"Id","Name","Username","Address","Private IP","TTL","State","Creation Time","Created By","Updated By")
	
			for _, g := range result {
				state := "N/A"
				if (len(g.Content) > 0) && (strings.Compare(g.Content,"N/A") != 0) {
					state = "open"
				} else {
					state = "close"
				}
				listRec(w,
					g.Id,
					g.Name,
					g.Username,
					g.Address,
					g.PrivateIp,
					g.Ttl,
					state,
					prettyTime{g.CreatedAt},
					g.CreatedBy,
					g.UpdatedBy,	
				)
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

	currentAccountId := findAccountId()
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
		printFatal("You should specify the name of gateway to open\ngateways open --name <gateway name> --key <path/to/gateway/key/file>  --ttl <time to live>")
		os.Exit(2)
	}

	accountId, gatewayId := findGatwayInfo(gatewayName)

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

	err = client.UpdateGateway(accountId, gatewayId, string(keyContent), flagTtl)

	if err != nil {
		printFatal("Error opening gateway : " + err.Error())
		os.Exit(2)
	}
	fmt.Println("Gateway opened successfully!")
}

func runRemoveGateway(c *cli.Context) {
	gatewayName := ""
	if c.IsSet("name") {
		gatewayName = c.String("name")
	} else {
		printFatal("You should specify the name of gateway to remove\ngateways remove --name <gateway name>")
		os.Exit(2)
	}

	accountId, gatewayId := findGatwayInfo(gatewayName)

	err := client.RemoveGateway(accountId, gatewayId)

	if err != nil {
		printFatal("Error remove gateway : " + err.Error())
		os.Exit(2)
	}
	fmt.Println("Gateway removed successfully!")
}

func runCloseGateway(c *cli.Context) {
	gatewayName := ""
	if c.IsSet("name") {
		gatewayName = c.String("name")
	} else {
		printFatal("You should specify the name of gateway to close\ngateways close --name <gateway name>")
		os.Exit(2)
	}

	accountId, gatewayId := findGatwayInfo(gatewayName)

	flagTtl := 1
	keyContent := ""

	err := client.UpdateGateway(accountId, gatewayId, string(keyContent), flagTtl)

	if err != nil {
		printFatal("Error close gateway : " + err.Error())
		os.Exit(2)
	}
	fmt.Println("Gateway closed successfully!")
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

func findGatwayInfo(gatewayName string) (int, int) {
	accountInfos, err := client.AccountInfos()
	if err != nil {
		printFatal("Can not retrive account info.")
		os.Exit(2)
	}
    	
	result_gateway_id := -1
	result_account_id := -1
	for _, accountInfo := range accountInfos {		
		gateways, err := client.ListGateways(accountInfo.Id)
		if err == nil {
			for _, g := range gateways {
				if strings.Compare(g.Name, gatewayName) == 0 {
					result_gateway_id = g.Id
					result_account_id = accountInfo.Id
					break
				}
			}
		}
		if result_gateway_id != -1 {
			break
		}		
	}

	if result_gateway_id == -1 {
		printFatal(fmt.Sprintf("Can not find gateway(%s)", gatewayName))
		os.Exit(2)
	}
	return result_account_id,result_gateway_id
}

package main

import (
	//	"fmt"
	"github.com/cloud66/cli"
	"io/ioutil"
	"os"
	"text/tabwriter"
)

var cmdGatewayKey = &Command{
	NeedsStack: false,
	Build:      buildGatewayKeys,
	Name:       "gateway-key",
	Short:      "commands to work with deploy gateway keys",
}

func buildGatewayKeys() cli.Command {
	base := buildBasicCommand()
	base.Subcommands = []cli.Command{
		cli.Command{
			Name:   "list",
			Action: runListGatewayKey,
			Usage:  "gateway-key list",
			Description: `lists gateway key of account
Examples:
$ cx gateway-key list
`,
		},
		cli.Command{
			Name:   "add",
			Action: runAddGatewayKey,
			Usage:  "gateway-key add [-t <time to live>] <path/to/gateway_key/file>",
			Description: `Add gateway key for an account.
-t allow you to specify the number of seconds that key will be available. If you did not specify ttl, the key will be available forever.
Examples:
$ cx gateway-key add  -t 1800  /tmp/gateway.pem
`,
			Flags: []cli.Flag{
				cli.IntFlag{
					Name: "ttl,t",
				},
			},
		},
		cli.Command{
			Name:   "remove",
			Action: runRemoveGatewayKey,
			Usage:  "gateway-key remove ",
			Description: `Remove gateway key from an account.

Example:
$ cx gateway-key remove`,
		},
	}

	return base
}

func runListGatewayKey(c *cli.Context) {
	currentAccountId := findAccountId()
	if currentAccountId == 0 {
		printFatal("Can not find current account")
		os.Exit(2)
	}

	gateway_key, err := client.ListGatewayKey(currentAccountId)
	if err != nil {
		printFatal("Error list gateway key : " + err.Error())
		os.Exit(2)
	}

	if gateway_key != nil {
		w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
		defer w.Flush()

		listRec(w,
			gateway_key.Id,
			gateway_key.Ttl,
			gateway_key.Content,
			gateway_key.CreatedAt,
		)
	}

}

func runAddGatewayKey(c *cli.Context) {
	if len(c.Args()) == 0 {
		printFatal("You should specify a key file")
		os.Exit(2)
	}
	key_path := c.Args()[0]

	flagTtl := 0

	if c.IsSet("ttl") {
		flagTtl = c.Int("ttl")
	}

	currentAccountId := findAccountId()
	if currentAccountId == 0 {
		printFatal("Can not find current account")
		os.Exit(2)
	}

	key_path_file := expandPath(key_path)
	key_content, err := ioutil.ReadFile(key_path_file)
	if err != nil {
		printFatal("Can not read from %s : "+err.Error(), key_path_file)
		os.Exit(2)
	}

	err = client.AddGatewayKey(currentAccountId, string(key_content), flagTtl)

	if err != nil {
		printFatal("Error add gateway key : " + err.Error())
		os.Exit(2)
	}
}

func runRemoveGatewayKey(c *cli.Context) {
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

/*
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
*/

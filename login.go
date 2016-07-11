package main

import (
	"fmt"
	"os"

	"github.com/cloud66/cli"
)

var (
	baseURL  = "https://app.cloud66.com"
	cmdLogin = &Command{
		Run:        runLogin,
		Name:       "login",
		Build:      buildBasicCommand,
		NeedsStack: false,
		NeedsOrg:   false,
		Short:      "opens the default web browser and logs into your Cloud 66 account",
		Long: `This makes it easier to login to Cloud 66 web UI by streamlining your login through a registered and authenticated CX.

Examples:
$ cx login 
`,
	}
)

func runLogin(c *cli.Context) {
	otp, err := client.AccountOTP()
	if err != nil {
		printFatal(err.Error())
	}

	baseAPIURL := os.Getenv("CLOUD66_API_URL")
	if baseAPIURL == "" {
		baseAPIURL = baseURL
	}

	toOpen := fmt.Sprintf("%s/otp?otp=%s", baseAPIURL, otp)
	err = openURL(toOpen)
	if err != nil {
		printFatal(err.Error())
	}

}

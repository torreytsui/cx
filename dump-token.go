package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/cloud66/cli"
)

var cmdDumpToken = &Command{
	Name:  "dump-token",
	Build: buildBasicCommand,
	Run:   runToken,
	Short: "prints the content of the cx token file with no new lines",
	Long: "The command can be used together with the 'base64' command to generate a base64 secret, which in turn " +
		"can be used with Github Actions.",
	NeedsStack: false,
	NeedsOrg:   false,
}

func runToken(*cli.Context) {

	tokenAbsolutePath := filepath.Join(cxHome(), profile.TokenFile)
	data, err := ioutil.ReadFile(tokenAbsolutePath)

	if err != nil {
		fmt.Println("File reading error: ", err)
		return
	}

	fmt.Print(strings.TrimSuffix(string(data), "\n")) // removes the new line, so that it can be converted in base64

}

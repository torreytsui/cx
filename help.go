package main

import "github.com/cloud66/cli"

var cmdHelpEnviron = &Command{
	Name:  "help-environ",
	Build: buildBasicCommand,
	Run:   dummyCommand,
	Short: "environment variables used by cx",
	Long: `
Several environment variables affect cx's behavior.

CXDEBUG
	When this is set, cx prints the wire representation of each API
  request to stderr just before sending the request, and prints the
  response. This will most likely include your secret API key in
  the Authorization header field, so be careful with the output.

CXTOKEN
	Anything set to this will be passed as X-CxToken HTTP header to the server.
	Used for development purposes.
`,
}

func dummyCommand(c *cli.Context) {
	cli.ShowCommandHelp(c, "help-environ")
}

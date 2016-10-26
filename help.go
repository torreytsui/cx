package main

import "github.com/cloud66/cli"

var cmdHelpEnviron = &Command{
	Name:  "help-environ",
	Build: buildBasicCommand,
	Run:   dummyCommand,
	Short: "environment variables used by cx",
	Long: `
Several environment variables affect cx's behavior.

CLOUD66_API_URL
	The base URL hk will use to make api requests in the format:
  https://host[:port]/

  Its default value is https://app.cloud66.com/api/2

CXDEBUG
	When this is set, cx prints the wire representation of each API
  request to stderr just before sending the request, and prints the
  response. This will most likely include your secret API key in
  the Authorization header field, so be careful with the output.

CXENVIRONMENT
	When this is set, cx looks for an environment specific OAuth token file.
	Normally the authentication file is called cx.json as is located under ~/.cloud66
	directory. Using this environment variable, cx will look for cx_<environment>.json
	file instead.

CX_APP_ID, CX_APP_SECRET
	Normally cx uses production environment app ID and secret. When this is set, the normal app ID
	and secrects are overwritten. This is used for debugging and development purposes.

CXSTACK
	If set, it overrides the stack sent as a parameter to the commands.
	This can be the name of the stack or its UID.

CXTOKEN
	Anything set to this will be passed as X-CxToken HTTP header to the server.
	Used for development purposes.
`,
}

func dummyCommand(c *cli.Context) {
	cli.ShowCommandHelp(c, "help-environ")
}

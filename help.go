package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"sort"
	"text/tabwriter"
)

var helpEnviron = &Command{
	Usage:    "environ",
	Category: "cx",
	Short:    "environment variables used by cx",
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

var cmdVersion = &Command{
	Run:      runVersion,
	Usage:    "version",
	Category: "cx",
	Short:    "show cx version",
	Long:     `Version shows the cx client version string.`,
}

func runVersion(cmd *Command, args []string) {
	fmt.Println(VERSION)
	if debugMode {
		fmt.Println("Running in debug mode")
		fmt.Printf("Build date: %s\n", BUILD_DATE)
	}
}

var cmdHelp = &Command{
	Usage:    "help [<topic>]",
	Category: "cx",
	Long:     `Help shows usage for a command or other topic.`,
}

var helpMore = &Command{
	Usage:    "more",
	Category: "cx",
	Short:    "additional commands, less frequently used",
	Long:     "(not displayed; see special case in runHelp)",
}

var helpCommands = &Command{
	Usage:    "commands",
	Category: "cx",
	Short:    "list all commands with usage",
	Long:     "(not displayed; see special case in runHelp)",
}

func init() {
	cmdHelp.Run = runHelp // break init loop
}

func runHelp(cmd *Command, args []string) {
	if len(args) == 0 {
		printUsageTo(os.Stdout)
		return // not os.Exit(2); success
	}
	if len(args) != 1 {
		printFatal("too many arguments")
	}
	switch args[0] {
	case helpMore.Name():
		printExtra()
		return
	case helpCommands.Name():
		printAllUsage()
		return
	}

	for _, cmd := range commands {
		if cmd.Name() == args[0] {
			cmd.printUsageTo(os.Stdout)
			return
		}
	}

	log.Printf("Unknown help topic: %q. Run 'cx help'.\n", args[0])
	os.Exit(2)
}

func maxStrLen(strs []string) (strlen int) {
	for i := range strs {
		if len(strs[i]) > strlen {
			strlen = len(strs[i])
		}
	}
	return
}

var usageTemplate = template.Must(template.New("usage").Parse(`
Usage: cx <command> [-s stack] [-e environment] [options] [arguments]


Commands:
{{range .Commands}}{{if .Runnable}}{{if .List}}
    {{.Name | printf (print "%-" $.MaxRunListName "s")}}  {{.Short}}{{end}}{{end}}{{end}}
{{range .Plugins}}
    {{.Name | printf (print "%-" $.MaxRunListName "s")}}  {{.Short}} (plugin){{end}}

Run 'cx help [command]' for details.


Additional help topics:
{{range .Commands}}{{if not .Runnable}}
    {{.Name | printf "%-8s"}}  {{.Short}}{{end}}{{end}}

{{if .Dev}}This dev build of cx cannot auto-update itself.
{{end}}`[1:]))

var extraTemplate = template.Must(template.New("usage").Parse(`
Additional commands:
{{range .Commands}}{{if .Runnable}}{{if .ListAsExtra}}
    {{.Name | printf (print "%-" $.MaxRunExtraName "s")}}  {{.ShortExtra}}{{end}}{{end}}{{end}}

Run 'cx help [command]' for details.

`[1:]))

func printUsageTo(w io.Writer) {
	var runListNames []string
	for i := range commands {
		if commands[i].Runnable() && commands[i].List() {
			runListNames = append(runListNames, commands[i].Name())
		}
	}

	usageTemplate.Execute(w, struct {
		Commands       []*Command
		Dev            bool
		MaxRunListName int
	}{
		commands,
		VERSION == "dev",
		maxStrLen(runListNames),
	})
}

func printExtra() {
	var runExtraNames []string
	for i := range commands {
		if commands[i].Runnable() && commands[i].ListAsExtra() {
			runExtraNames = append(runExtraNames, commands[i].Name())
		}
	}

	extraTemplate.Execute(os.Stdout, struct {
		Commands        []*Command
		MaxRunExtraName int
	}{
		commands,
		maxStrLen(runExtraNames),
	})
}

func printAllUsage() {
	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()
	cl := commandList(commands)
	sort.Sort(cl)
	for i := range cl {
		if cl[i].Runnable() {
			listRec(w, "cx "+cl[i].FullUsage(), "# "+cl[i].Short)
		}
	}
}

type commandList []*Command

func (cl commandList) Len() int           { return len(cl) }
func (cl commandList) Swap(i, j int)      { cl[i], cl[j] = cl[j], cl[i] }
func (cl commandList) Less(i, j int) bool { return cl[i].Name() < cl[j].Name() }

type commandMap map[string]commandList

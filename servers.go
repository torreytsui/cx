package main

import (
	"io"
	"os"
	"sort"
	"strings"

	"text/tabwriter"

	"github.com/cloud66/cloud66"
)

var cmdServers = &Command{
	Run:        runServers,
	Usage:      "servers [<names>]",
	NeedsStack: true,
	Category:   "stack",
	Short:      "lists all the servers of a stack.",
	Long: `List all the servers of a stack.
The information contains the name, IP address, server role and the date/time it was created.
Names can be a list of server names entered as separate parameters.

Examples:
$ cx servers -s mystack
orca         162.243.201.164  [rails web app]  Healthy   Mar 26 11:23
snowleopard  107.170.98.160   [mysql db]       Building  Mar 26 11:23
$ cx servers -s mystack orca
orca         162.243.201.164  [rails web app]  Healthy   Mar 26 11:23
`,
}

func runServers(cmd *Command, serverNames []string) {
	stack := mustStack()
	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()

	var servers []cloud66.Server
	var err error
	servers, err = client.Servers(stack.Uid)
	must(err)

	for idx, i := range serverNames {
		serverNames[idx] = strings.ToLower(i)
	}
	sort.Strings(serverNames)
	if len(serverNames) == 0 {
		printServerList(w, servers)
	} else {
		// filter out the unwanted servers
		var filteredServers []cloud66.Server
		for _, i := range servers {
			sorted := sort.SearchStrings(serverNames, strings.ToLower(i.Name))
			if sorted < len(serverNames) && strings.ToLower(serverNames[sorted]) == strings.ToLower(i.Name) {
				filteredServers = append(filteredServers, i)
			}
		}
		printServerList(w, filteredServers)
	}
}

func printServerList(w io.Writer, servers []cloud66.Server) {
	sort.Sort(serversByName(servers))
	for _, a := range servers {
		if a.Name != "" {
			listServer(w, a)
		}
	}
}

func listServer(w io.Writer, a cloud66.Server) {
	t := a.CreatedAt
	listRec(w,
		strings.ToLower(a.Name),
		a.Address,
		a.Roles,
		a.Health(),
		prettyTime{t},
	)
}

type serversByName []cloud66.Server

func (a serversByName) Len() int           { return len(a) }
func (a serversByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a serversByName) Less(i, j int) bool { return a[i].Name < a[j].Name }

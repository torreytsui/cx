package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/cloud66/cloud66"

	"github.com/codegangsta/cli"
)

var cmdEasyDeploy = &Command{
	Name:       "easydeploys",
	Build:      buildEasyDeply,
	Short:      "commands to work with EasyDeploy",
	NeedsStack: false,
}

func buildEasyDeply() cli.Command {
	base := buildBasicCommand()

	base.Subcommands = []cli.Command{
		cli.Command{
			Name:   "list",
			Action: runEasyDeploy,
			Usage:  "list or shows information about available remote EasyDeploy apps available",
			Description: `This lists all the available EasyDeploy apps or detailed information about a specific one.

Examples:
$ cx easydeploys list
wordpress
gitlab

$ cx easydeploys list wordpress
wordpress  WordPress  4.1.0  2015-01-19
`,
		},
	}

	return base
}

func runEasyDeploy(c *cli.Context) {
	appNames := c.Args()

	if len(appNames) != 0 {
		w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
		defer w.Flush()

		var items []cloud66.EasyDeploy
		for _, appName := range appNames {
			item, err := client.EasyDeployInfo(appName)
			if err != nil {
				printFatal(err.Error())
				return
			}

			items = append(items, *item)
		}

		printEasyDeployList(w, items)

	} else {
		list, err := client.EasyDeployList()
		if err != nil {
			printFatal("Error during fetching the list: " + err.Error())
			return
		}

		for _, easyDeploy := range list {
			fmt.Println(easyDeploy)
		}
	}
}

func printEasyDeployList(w io.Writer, easyDeploys []cloud66.EasyDeploy) {
	sort.Sort(easyDeploysByName(easyDeploys))
	for _, a := range easyDeploys {
		listEasyDeploy(w, a)
	}
}

func listEasyDeploy(w io.Writer, a cloud66.EasyDeploy) {
	listRec(w,
		a.Name,
		*a.DisplayName,
		a.Version,
		a.CreatedAt,
	)
}

type easyDeploysByName []cloud66.EasyDeploy

func (a easyDeploysByName) Len() int           { return len(a) }
func (a easyDeploysByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a easyDeploysByName) Less(i, j int) bool { return a[i].Name < a[j].Name }

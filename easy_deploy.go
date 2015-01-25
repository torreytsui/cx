// +build ignore

package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/cloud66/cloud66"
)

var cmdEasyDeploy = &Command{
	Run:        runEasyDeploy,
	Usage:      "easydeploys [app_name,...]",
	NeedsStack: false,
	Category:   "easydeploy",
	Short:      "list or shows information about available remote EasyDeploy apps available",
	Long: `This lists all the available EasyDeploy apps or detailed information about a specific one.

Examples:
$ cx easydeploys
wordpress
gitlab

$ cx easydeploys wordpress
wordpress  WordPress  4.1.0  2015-01-19
`,
}

func runEasyDeploy(cmd *Command, appNames []string) {
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

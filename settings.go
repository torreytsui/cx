package main

import (
	"io"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/cloud66/cloud66"
)

var cmdSettings = &Command{
	Run:        runSettings,
	Usage:      "settings [settings]",
	NeedsStack: true,
	Category:   "stack",
	Short:      "lists stack settings",
	Long: `Lists all the settings applicable to the given stack.
  It also shows the key, value and the readonly flag for each setting.
  Settings can be a list of multiple settings as separate parameters.
  To change each setting, use the set command.

Examples:

    $ cx settings -s mystack
    git.branch          master                                                     false
    git.repository      git://git@github.com:cloud66-samples/rails-psql-redis.git  false
    allowed.web.source  <nil>                                                      false

    $ cx settings -s mystack git.branch
    git.branch          master                                                     false
`,
}

func runSettings(cmd *Command, settingNames []string) {
	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()
	var settings []cloud66.StackSetting
	var err error
	stack := mustStack()
	settings, err = client.StackSettings(stack.Uid)
	must(err)

	sort.Strings(settingNames)
	if len(settingNames) == 0 {
		printSettingList(w, settings)
	} else {
		// filter out the unwanted settings
		var filteredSettings []cloud66.StackSetting
		for _, i := range settings {
			sorted := sort.SearchStrings(settingNames, i.Key)
			if sorted < len(settingNames) && settingNames[sorted] == i.Key {
				filteredSettings = append(filteredSettings, i)
			}
		}

		printSettingList(w, filteredSettings)
	}
}

func printSettingList(w io.Writer, settings []cloud66.StackSetting) {
	sort.Sort(settingsByName(settings))
	for _, a := range settings {
		if a.Key != "" {
			listSetting(w, a)
		}
	}
}

func listSetting(w io.Writer, a cloud66.StackSetting) {
	var readonly string
	if a.Readonly {
		readonly = "readonly"
	} else {
		readonly = "read/write"
	}
	listRec(w,
		a.Key,
		a.Value,
		readonly,
	)
}

type settingsByName []cloud66.StackSetting

func (a settingsByName) Len() int           { return len(a) }
func (a settingsByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a settingsByName) Less(i, j int) bool { return a[i].Key < a[j].Key }

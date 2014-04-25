package main

import (
	"io"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/cloud66/cx/cloud66"
)

var cmdEnvVars = &Command{
	Run:        runEnvVars,
	Usage:      "env-vars [environment_variables]",
	NeedsStack: true,
	Category:   "stack",
	Short:      "lists environement variables",
	Long: `Lists all the environement variables of the given stack.  
  The environment_variables options can be a list of multiple environment_variables as separate parameters.
  To change environement variable values, use the env-vars-set command.

Examples:

    $ cx env-vars -s mystack
    RAILS_ENV 			production
    STACK_BASE      	/abc/def
    STACK_PATH      	/abc/def/current
    etc..

    $ cx env-vars -s mystack RAILS_ENV
    RAILS_ENV 			production

    $ cx env-vars -s mystack RAILS_ENV STACK_BASE 
    RAILS_ENV 			production
    STACK_BASE      	/abc/def
`,
}

func runEnvVars(cmd *Command, envVarKeys []string) {
	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()	
	var envVars []cloud66.StackEnvVar
	var err error
	stack := mustStack()
	envVars, err = client.StackEnvVars(stack.Uid)
	must(err)

	sort.Strings(envVarKeys)
	if len(envVarKeys) == 0 {
		printEnvVarsList(w, envVars)
	} else {
		// filter out the unwanted env_vars
		var filteredEnvVars []cloud66.StackEnvVar
		for _, i := range envVars {
			sorted := sort.SearchStrings(envVarKeys, i.Key)
			if sorted < len(envVarKeys) && envVarKeys[sorted] == i.Key {
				filteredEnvVars = append(filteredEnvVars, i)
			}
		}

		printEnvVarsList(w, filteredEnvVars)
	}
}

func printEnvVarsList(w io.Writer, envVars []cloud66.StackEnvVar) {
	sort.Sort(envVarsByName(envVars))
	for _, a := range envVars {
		if a.Key != "" {
			listEnvVar(w, a)
		}
	}
}

func listEnvVar(w io.Writer, a cloud66.StackEnvVar) {
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

type envVarsByName []cloud66.StackEnvVar

func (a envVarsByName) Len() int           { return len(a) }
func (a envVarsByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a envVarsByName) Less(i, j int) bool { 
	if a[i].Readonly == a[j].Readonly {
		return a[i].Key < a[j].Key 
	}
	return boolToInt(a[i].Readonly) < boolToInt(a[j].Readonly)
}

func boolToInt(b bool) int {
    if b {
        return 1
    }
    return 0
}

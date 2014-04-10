package main

import (
	"io"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/cloud66/cx/cloud66"
)

var cmdStacks = &Command{
	Run:      runStacks,
	Usage:    "stacks [<names>] [-e <environment>]",
	Category: "stack",
	Short:    "list stacks",
	Long: `
Lists stacks. Shows the stack name, environment, and last deploy time.
You can use multiple names at the same time.

Examples:

    $ cx stacks
    mystack     production   Jan 2 12:34
    mystack     staging      Feb 2 12:34
    mystack-2   development  Jan 2 12:35

    $ cx stacks mystack-2
    mystack-2   development  Jan 2 12:34

    $ cx stacks mystack -e staging
    mystack     staging      Feb 2 12:34
`,
}

var
	flagForcedEnvironment string

func init() {
	cmdStacks.Flag.StringVar(&flagForcedEnvironment, "e", "", "stack environment")
}

func runStacks(cmd *Command, names []string) {
	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()
	var stacks []cloud66.Stack
	if len(names) == 0 {
		var err error
		stacks, err = client.StackListWithFilter(func (item interface{}) bool {
			if flagForcedEnvironment == "" {
				return true
			}

			return strings.HasPrefix(strings.ToLower(item.(cloud66.Stack).Environment), strings.ToLower(flagForcedEnvironment))
		})
		must(err)
	} else {
		stackch := make(chan *cloud66.Stack, len(names))
		errch := make(chan error, len(names))
		for _, name := range names {
			if name == "" {
				stackch <- nil
			} else {
				go func(stackname string) {
					if stack, err := client.StackInfoWithEnvironment(stackname, flagForcedEnvironment); err != nil {
						errch <- err
					} else {
						stackch <- stack
					}
				}(name)
			}
		}
		for _ = range names {
			select {
			case err := <-errch:
				printFatal(err.Error())
			case stack := <-stackch:
				if stack != nil {
					stacks = append(stacks, *stack)
				}
			}
		}
	}
	printStackList(w, stacks)
}

func printStackList(w io.Writer, stacks []cloud66.Stack) {
	sort.Sort(stacksByName(stacks))
	for _, a := range stacks {
		if a.Name != "" {
			listStack(w, a)
		}
	}
}

func listStack(w io.Writer, a cloud66.Stack) {
	t := a.CreatedAt
	if a.LastActivity != nil {
		t = *a.LastActivity
	}
	listRec(w,
		a.Name,
		a.Environment,
		a.Status(),
		prettyTime{t},
	)
}

type stacksByName []cloud66.Stack

func (a stacksByName) Len() int           { return len(a) }
func (a stacksByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a stacksByName) Less(i, j int) bool { return a[i].Name < a[j].Name }

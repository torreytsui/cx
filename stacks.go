package main

import (
	"io"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/cloud66/cloud66"

	"github.com/cloud66/cli"
)

var cmdStacks = &Command{
	Name:  "stacks",
	Build: buildStacks,
	Short: "commands to work with stacks",
}

func buildStacks() cli.Command {
	base := buildBasicCommand()
	base.Subcommands = []cli.Command{
		cli.Command{
			Name:  "list",
			Usage: "lists all stacks",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "environment,e",
					Usage: "Full or partial environment name.",
				},
			},
			Action: runStacks,
			Description: `Lists stacks. Shows the stack name, environment, and last deploy time.
You can use multiple names at the same time.

Examples:
$ cx stacks list
mystack     production   Jan 2 12:34
mystack     staging      Feb 2 12:34
mystack-2   development  Jan 2 12:35

$ cx stacks list mystack-2
mystack-2   development  Jan 2 12:34

$ cx stacks list mystack -e staging
mystack     staging      Feb 2 12:34
`,
		},
		cli.Command{
			Name:  "create",
			Usage: "creates new docker stack",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "name,n",
					Usage: "New docker stack name.",
				},
				cli.StringFlag{
					Name:  "environment,e",
					Usage: "New docker stack environment.",
				},
				cli.StringFlag{
					Name:  "service_yaml,s",
					Usage: "File containing your service definition.",
				},
				cli.StringFlag{
					Name:  "manifest_yaml,m",
					Usage: "File containing your manifest definition (optional)",
				},
			},
			Action: runCreateStack,
			Description: `Creates a new docker stack.

Examples:
$ cx stacks create --name my_docker_stack --environment production --service_yaml service.yml --manifest_yaml manifest.yml
`,
		},
		cli.Command{
			Name:  "redeploy",
			Usage: "redeploys a stack",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "y",
					Usage: "answer yes to confirmations",
				},
				cli.StringFlag{
					Name:  "git-ref",
					Usage: "git reference",
				},
				cli.StringFlag{
					Name:  "services",
					Usage: "comma separated list of services to include in the deploy",
				},
				cli.StringFlag{
					Name:  "environment,e",
					Usage: "Full or partial environment name.",
				},
				cli.StringFlag{
					Name:  "stack,s",
					Usage: "Full or partial stack name. This can be omited if the current directory is a stack directory",
				},
			},
			Action: runRedeploy,
			Description: `Enqueues redeployment of the stack.
If the stack is already building, another build will be enqueued and performed immediately
after the current one is finished.

-y answers yes to confirmation question if the stack is production.
--git-ref will redeploy the specific branch, tag or hash git reference (non-docker stacks)
--services will deploy the specified services from your stack only (docker stacks)
`,
		},
		cli.Command{
			Name:   "restart",
			Action: runRestart,
			Flags:  basicFlags(),
			Usage:  "restarts all components of a stack",
			Description: `This will send a restart method to all stack components. This means different things for different components.
For a web server, it means a restart of nginx. For an application server, this might be a restart of the workers like Unicorn.
For more information on restart command, please refer to help.cloud66.com
`,
		},
		cli.Command{
			Name:   "clear-caches",
			Action: runClearCaches,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "environment,e",
					Usage: "Full or partial environment name.",
				},
				cli.StringFlag{
					Name:  "stack,s",
					Usage: "Full or partial stack name. This can be omited if the current directory is a stack directory",
				},
			},
			Usage: "clears all existing stack code caches",
			Description: `Clears all existing code caches.

For improved performance, volatile code caches exist for your stack.
It is possible for a those volatile caches to become invalid if you switch branches, change git repository URL, or rebase or force a commit.
Since switching branch or changing git repository URL is done via the Cloud 66 interface, your volatile caches will automatically be purged.
However, rebasing or forcing a commit doesn't have any association with Cloud 66, so this command can be used to purge the exising volatile caches.
`},
		cli.Command{
			Name:   "listen",
			Action: runListen,
			Usage:  "tails all deployment logs",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "environment,e",
					Usage: "Full or partial environment name.",
				},
				cli.StringFlag{
					Name:  "stack,s",
					Usage: "Full or partial stack name. This can be omited if the current directory is a stack directory",
				},
			},
			Description: `This acts as a log tail for deployment of a stack so you don't have to follow the deployment on the web.

Examples:
$ cx stacks listen
$ cx stacks listen -s mystack
`},
	}

	return base
}

func runStacks(c *cli.Context) {
	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()
	var stacks []cloud66.Stack
	names := c.Args()
	flagForcedEnvironment := c.String("environment")
	if len(names) == 0 {
		var err error
		stacks, err = client.StackListWithFilter(func(item interface{}) bool {
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

func basicFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "environment,e",
			Usage: "Full or partial environment name.",
		},
		cli.StringFlag{
			Name:  "stack,s",
			Usage: "Full or partial stack name. This can be omited if the current directory is a stack directory",
		},
	}
}

type stacksByName []cloud66.Stack

func (a stacksByName) Len() int           { return len(a) }
func (a stacksByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a stacksByName) Less(i, j int) bool { return a[i].Name < a[j].Name }

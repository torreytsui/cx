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
					Usage: "full or partial environment name",
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
					Usage: "[classic stacks] git reference",
				},
				cli.StringSliceFlag{
					Name:  "service",
					Usage: "[docker stacks] service name (and optional colon separated reference) to include in the deploy. Repeatable for multiple services",
					Value: &cli.StringSlice{},
				},
				cli.BoolFlag{
					Name:  "listen",
					Usage: "show stack deployment progress and log output",
				},
				cli.StringFlag{
					Name:  "environment,e",
					Usage: "full or partial environment name",
				},
				cli.StringFlag{
					Name:  "stack,s",
					Usage: "full or partial stack name. This can be omitted if the current directory is a stack directory",
				},
			},
			Action: runRedeploy,
			Description: `Enqueues redeployment of the stack.
If the stack is already building, another build will be enqueued and performed immediately
after the current one is finished.

-y answers yes to confirmation question if the stack is production.
--git-ref will redeploy the specific branch, tag or hash git reference [classic stacks]
--service is a repeateable option to deploy only the specified service(s). Including a reference (separated by a colon) will attempt to deploy that particular reference for that service [docker stacks]
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
			Name:   "reboot",
			Action: runStackReboot,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "y",
					Usage: "answer yes to confirmations",
				},
				cli.StringFlag{
					Name:  "group",
					Usage: "Specify which group you would like to reboot",
				},
				cli.StringFlag{
					Name:  "strategy",
					Usage: "Specify how you would like to reboot your servers",
				},
				cli.StringFlag{
					Name:  "environment,e",
					Usage: "full or partial environment name",
				},
				cli.StringFlag{
					Name:  "stack,s",
					Usage: "full or partial stack name. This can be omitted if the current directory is a stack directory",
				},
			},
			Usage: "reboot servers in your stack",
			Description: `reboot servers in your stack.

The group parameter specifies which group of servers you wish to reboot. Valid values are "all", "web", "haproxy", "db";
DB specific values like "mysql" or "redis" for example are also supported.
If this value is left unspecified, the default value of "web" will be used

The strategy parameter specifies whether you want all your servers to be rebooted in parallel or in serial.
Valid values for this parameter are "serial" or "parallel"; "serial" reboots involves web servers being removed/re-added to the LB one by one.
Note that for this only applies to web servers; non-web server will still be rebooted in parallel.
If this value is left unspecified, Cloud 66 will determine the best strategy based on your infrastructure layout.

Examples:
$ cx stack reboot -s mystack
$ cx stack reboot -s mystack --group web
$ cx stack reboot -s mystack --group all
$ cx stack reboot -s mystack --strategy parallel
$ cx stack reboot -s mystack --group web --strategy serial 
`},

		cli.Command{
			Name:   "clear-caches",
			Action: runClearCaches,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "environment,e",
					Usage: "full or partial environment name",
				},
				cli.StringFlag{
					Name:  "stack,s",
					Usage: "full or partial stack name. This can be omitted if the current directory is a stack directory",
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
					Usage: "full or partial environment name",
				},
				cli.StringFlag{
					Name:  "stack,s",
					Usage: "full or partial stack name. This can be omitted if the current directory is a stack directory",
				},
			},
			Description: `This acts as a log tail for deployment of a stack so you don't have to follow the deployment on the web.

Examples:
$ cx stacks listen
$ cx stacks listen -s mystack
`},
		cli.Command{
			Name:  "configure",
			Usage: "list, download and upload of configuration files",
			Subcommands: []cli.Command{
				cli.Command{
					Name:   "list",
					Action: runStackConfigureFileList,
					Usage:  "list of all versions of a configuration file",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "file,f",
							Usage: "supported values are: service.yml , manifest.yml",
						},
						cli.StringFlag{
							Name:  "environment,e",
							Usage: "full or partial environment name",
						},
						cli.StringFlag{
							Name:  "stack,s",
							Usage: "full or partial stack name. This can be omitted if the current directory is a stack directory",
						},
					},
					Description: `This acts list of all versions of configuration file.
`},
				cli.Command{
					Name:   "download",
					Action: runStackConfigureDownloadFile,
					Usage:  "download a configuration file",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "file,f",
							Usage: "supported values are: service.yml , manifest.yml",
						},
						cli.StringFlag{
							Name:  "version,v",
							Usage: "full or partial file version (optional)",
						},
						cli.StringFlag{
							Name:  "output,o",
							Usage: "full path of output file (optional)",
						},
						cli.StringFlag{
							Name:  "environment,e",
							Usage: "full or partial environment name",
						},
						cli.StringFlag{
							Name:  "stack,s",
							Usage: "full or partial stack name. This can be omitted if the current directory is a stack directory",
						},
					},
					Description: `download service.yml and manifest.yml.
`},
				cli.Command{
					Name:   "upload",
					Action: runStackConfigureUploadFile,
					Usage:  "uploading new version of configuration file",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "file,f",
							Usage: "supported values are: service.yml , manifest.yml",
						},
						cli.StringFlag{
							Name:  "comments,c",
							Usage: "a brief description of your changes",
						},
						cli.StringFlag{
							Name:  "environment,e",
							Usage: "full or partial environment name",
						},
						cli.StringFlag{
							Name:  "stack,s",
							Usage: "full or partial stack name. This can be omitted if the current directory is a stack directory",
						},
					},
					Description: `upload new service.yml or manifest.yml.
`},
			},
			Description: `

Examples:
$ cx stacks configure list -f service.yml -s mystack
$ cx stacks configure download -f manifest.yml -s mystack
$ cx stacks configure download -f service.yml -o /tmp/my_stack_servive.yml -s mystack
$ cx stacks configure download -f manifest.yml -v f345 -s mystack
$ cx stacks configure upload /tmp/mystack_edited_service.yml -f service.yml -s mystack --comments "new service added"
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
			Usage: "full or partial environment name",
		},
		cli.StringFlag{
			Name:  "stack,s",
			Usage: "full or partial stack name. This can be omitted if the current directory is a stack directory",
		},
	}
}

type stacksByName []cloud66.Stack

func (a stacksByName) Len() int           { return len(a) }
func (a stacksByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a stacksByName) Less(i, j int) bool { return a[i].Name < a[j].Name }

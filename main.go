package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/cloud66/cloud66"
	//	"github.com/cloud66/cx/term"

	"github.com/codegangsta/cli"
	"github.com/jcoene/honeybadger"
	//	"github.com/mgutz/ansi"
)

type Command struct {
	Name       string
	Build      func() cli.Command
	Run        func(c *cli.Context)
	Flags      []cli.Flag
	Short      string
	Long       string
	NeedsStack bool
}

var (
	client     cloud66.Client
	debugMode  bool   = false
	VERSION    string = "dev"
	BUILD_DATE string = ""
	tokenFile  string = "cx.json"
	nsqLookup  string = "nsq.cldblx.com:4161"
)

var commands = []*Command{
	cmdStacks,
	cmdOpen,
	cmdSettings,
	//cmdEasyDeploy,
	cmdEnvVars,
	cmdLease,
	cmdListen,
	//		cmdRun,
	cmdServers,
	cmdSsh,
	cmdTail,
	cmdUpload,
	cmdDownload,
	cmdBackups,
	/*
		cmdContainers,
		cmdContainerStop,
		cmdContainerRestart,*/
	cmdServices,
	/*
		cmdSlavePromote,
		cmdSlaveResync,

		cmdVersion,
		cmdUpdate,
		cmdHelp,*/
	cmdInfo,

	/*helpCommands,
	helpEnviron,
	helpMore,*/
}

var (
	flagStack       *cloud66.Stack
	flagEnvironment string
)

func main() {
	honeybadger.ApiKey = "09d82034"
	defer recoverPanic()

	app := cli.NewApp()
	app.Name = "cx"
	app.Usage = "Cloud 66 Command line toolbelt"
	app.Author = "Cloud 66"
	app.Email = "support@cloud66.com"
	app.Action = doMain
	app.Version = VERSION

	cmds := []cli.Command{}
	for _, cmd := range commands {
		cliCommand := cmd.Build()
		if cmd.Name == "" {
			printFatal("No Name is specified for %s", cmd)
		}
		cliCommand.Name = cmd.Name
		cliCommand.Usage = cmd.Short
		cliCommand.Description = cmd.Long
		cliCommand.Action = cmd.Run
		cliCommand.Flags = cmd.Flags

		if len(cliCommand.Subcommands) == 0 {
			if cmd.NeedsStack {
				cliCommand.Flags = append(cliCommand.Flags, cli.StringFlag{
					Name:  "stack,s",
					Usage: "Full or partial stack name. This can be omited if the current directory is a stack directory",
				}, cli.StringFlag{
					Name:  "environment,e",
					Usage: "Full or partial environment name.",
				})
			}
		} else {
			for idx, sub := range cliCommand.Subcommands {
				if cmd.NeedsStack {
					sub.Flags = append(sub.Flags, cli.StringFlag{
						Name:  "stack,s",
						Usage: "Full or partial stack name. This can be omited if the current directory is a stack directory",
					}, cli.StringFlag{
						Name:  "environment,e",
						Usage: "Full or partial environment name.",
					})
				}

				cliCommand.Subcommands[idx].Flags = sub.Flags
			}
		}

		cmds = append(cmds, cliCommand)
	}

	app.Commands = cmds

	setGlobals(app)

	app.Run(os.Args)
}

func setGlobals(app *cli.App) {
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "runenv",
			Usage: "sets the environment this toolbelt is running agains",
			//ShowHelp: false,
			Value:  "production",
			EnvVar: "CXENVIRONMENT",
		},
		cli.StringFlag{
			Name:  "nsqlookup",
			Usage: "sets the NSQ lookup address this toolbelt is running against",
			//HideHelp: false,
			EnvVar: "NSQ_LOOKUP",
		},
		cli.StringFlag{
			Name:  "cxstack",
			Usage: "CXSTACK",
			//ShowHelp: false,
			EnvVar: "CXSTACK",
		},
		cli.BoolFlag{
			Name:   "debug",
			Usage:  "run in debug more",
			EnvVar: "CXDEBUG",
		},
	}

	if os.Getenv("CXENVIRONMENT") != "" {
		tokenFile = "cx_" + os.Getenv("CXENVIRONMENT") + ".json"
		fmt.Printf("Running against %s environment\n", os.Getenv("CXENVIRONMENT"))
		honeybadger.Environment = os.Getenv("CXENVIRONMENT")
	} else {
		honeybadger.Environment = "production"
	}

	if os.Getenv("NSQ_LOOKUP") != "" {
		nsqLookup = os.Getenv("NSQ_LOOKUP")
	}

	// TODO: this is temp until we add an init command
	if len(os.Args) != 2 || os.Args[1] != "--version" {
		initClients()
	}
}

func buildBasicCommand() cli.Command {
	return cli.Command{}
}

func doMain(c *cli.Context) {
	/*
				if args[0] == cmdUpdate.Name() {
					cmdUpdate.Run(cmdUpdate, args[1:])
					return
				} else if VERSION != "dev" {
					defer backgroundRun()
				}
				if !term.IsANSI(os.Stdout) {
					ansi.DisableColors(true)
				}

				// don't need registration if we are only checking the version
				if args[0] != "version" {
					initClients()
				}
			for _, cmd := range commands {

				if cmd.Name() == args[0] && cmd.Run != nil {
					defer recoverPanic()

					cmd.Flag.Usage = func() {
						cmd.printUsage()
					}
					if cmd.NeedsStack {
						cmd.Flag.StringVar(&flagStackName, "s", "", "stack name")
						cmd.Flag.StringVar(&flagEnvironment, "e", "", "stack environment")
					}
					// optional server/servicename flag used in multiple places
					cmd.Flag.StringVar(&flagServer, "server", "", "server filter")
					cmd.Flag.StringVar(&flagServiceName, "service", "", "service name")
					cmd.Flag.StringVar(&flagDbType, "db-type", "", "database type")

					if err := cmd.Flag.Parse(args[1:]); err != nil {
						os.Exit(2)
					}
					if cmd.NeedsStack {
						// by default print server output to stdout
						var toSdout bool = true

						// when command is 'run', do not print server output to stdout
						if args[0] == "run" {
							toSdout = false
						}

						s, err := stack(toSdout)
						switch {
						case err == nil && s == nil:
							msg := "no stack specified"
							if err != nil {
								msg = err.Error()
							}
							printError(msg)
							cmd.printUsage()
							os.Exit(2)
						case err != nil:
							printFatal(err.Error())
						}
					}
					cmd.Run(cmd, cmd.Flag.Args())
					return
				}
			}

		// invalid command
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", args[0])
		if g := suggest(args[0]); len(g) > 0 {
			fmt.Fprintf(os.Stderr, "Possible alternatives: %v\n", strings.Join(g, " "))
		}
		fmt.Fprintf(os.Stderr, "Run 'cx help' for usage.\n")
		os.Exit(2)*/
}

func initClients() {
	// is there a token file?
	_, err := os.Stat(filepath.Join(cxHome(), tokenFile))
	if err != nil {
		fmt.Println("No previous authentication found.")
		cloud66.Authorize(cxHome(), tokenFile)
		os.Exit(1)
	} else {
		client = cloud66.GetClient(cxHome(), tokenFile, VERSION)
		debugMode = os.Getenv("CXDEBUG") != ""
		client.Debug = debugMode
	}
}

func recoverPanic() {
	if VERSION != "dev" {
		if rec := recover(); rec != nil {
			report, err := honeybadger.NewReport(rec)
			if err != nil {
				printError("reporting crash failed: %s", err.Error())
				panic(rec)
			}
			report.AddContext("Version", VERSION)
			report.AddContext("Platform", runtime.GOOS)
			report.AddContext("Architecture", runtime.GOARCH)
			report.AddContext("DebugMode", debugMode)
			result := report.Send()
			if result != nil {
				printError("reporting crash failed: %s", result.Error())
				panic(rec)
			}
			printFatal("cx encountered and reported an internal client error")
		}
	}
}

func filterByEnvironment(item interface{}) bool {
	if flagEnvironment == "" {
		return true
	}

	return strings.HasPrefix(strings.ToLower(item.(cloud66.Stack).Environment), strings.ToLower(flagEnvironment))
}

func stack(c *cli.Context) (*cloud66.Stack, error) {
	if flagStack != nil {
		return flagStack, nil
	}

	if c.String("environment") != "" {
		flagEnvironment = c.String("environment")
	}

	var err error
	if c.String("stack") != "" {
		stacks, err := client.StackListWithFilter(filterByEnvironment)
		if err != nil {
			return nil, err
		}
		var stackNames []string
		for _, stack := range stacks {
			stackNames = append(stackNames, stack.Name)
		}
		idx, err := fuzzyFind(stackNames, c.String("stack"), false)
		if err != nil {
			return nil, err
		}

		flagStack = &stacks[idx]

		// toSdout is of type []bool. Take first value
		if c.String("environment") != "" {
			fmt.Printf("(%s)\n", flagStack.Environment)
		}

		return flagStack, err
	}

	if stack := c.String("cxstack"); stack != "" {
		// the environment variable should be exact match
		flagStack, err = client.StackInfo(stack)
		return flagStack, err
	}

	return stackFromGitRemote(remoteGitUrl(), localGitBranch())
}

func mustStack(c *cli.Context) *cloud66.Stack {
	stack, err := stack(c)
	if err != nil {
		printFatal(err.Error())
	}

	if stack == nil {
		printFatal("No stack specified. Either use --stack flag to cd to a stack directory")
	}

	return stack
}

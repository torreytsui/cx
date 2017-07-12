package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/cloud66/cli"
	"github.com/cloud66/cloud66"
	"github.com/getsentry/raven-go"
	"github.com/mgutz/ansi"
)

type Command struct {
	Name       string
	Build      func() cli.Command
	Run        func(c *cli.Context)
	Flags      []cli.Flag
	Short      string
	Long       string
	NeedsStack bool
	NeedsOrg   bool
}

const (
	redirectURL = "urn:ietf:wg:oauth:2.0:oob"
	scope       = "public redeploy jobs users admin"
)

var (
	client       cloud66.Client
	clientId     string
	clientSecret string
	debugMode    bool   = false
	underTest    bool   = false
	VERSION      string = "dev"
	BUILD_DATE   string = ""
	tokenFile    string = "cx.json"
	fayeEndpoint string = "https://sockets.cloud66.com:443/push"
)

var commands = []*Command{
	cmdStacks,
	cmdLogin,
	cmdRedeploy,
	cmdOpen,
	cmdSettings,
	cmdEasyDeploy,
	cmdEnvVars,
	cmdLease,
	cmdRun,
	cmdTunnel,
	cmdServers,
	cmdSsh,
	cmdTail,
	cmdUpload,
	cmdDownload,
	cmdBackups,
	cmdContainers,
	cmdServices,
	cmdDatabases,
	cmdJobs,
	cmdHelpEnviron,
	cmdUpdate,
	cmdInfo,
	cmdTest,
	cmdUsers,
	cmdGateway,
	cmdProcesses,
	cmdRegisterServer,
	cmdVersion,
}

var (
	flagStack       *cloud66.Stack
	flagOrg         *cloud66.Account
	flagEnvironment string
)

func main() {
	// add aliases for commands
	commands = populateAliases(commands)

	raven.SetDSN("https://39c187859231424fb4865e90d42a29a3:cfbc35db1b954f04be995a3d0ec3fbae@sentry.io/153008")
	defer recoverPanic()

	app := cli.NewApp()

	cmds := []cli.Command{}
	cli.VersionPrinter = runVersion

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
				cliCommand.Flags = append(cliCommand.Flags,
					cli.StringFlag{
						Name:  "stack,s",
						Usage: "full or partial stack name. This can be omitted if the current directory is a stack directory",
					}, cli.StringFlag{
						Name:  "environment,e",
						Usage: "full or partial environment name",
					})
			}
			if cmd.NeedsOrg {
				cliCommand.Flags = append(cliCommand.Flags,
					cli.StringFlag{
						Name:  "org",
						Usage: "full or partial organization name.",
					})
			}
		} else {
			for idx, sub := range cliCommand.Subcommands {
				if cmd.NeedsStack {
					sub.Flags = append(sub.Flags,
						cli.StringFlag{
							Name:  "stack,s",
							Usage: "full or partial stack name. This can be omitted if the current directory is a stack directory",
						}, cli.StringFlag{
							Name:  "environment,e",
							Usage: "full or partial environment name",
						})
				}
				if cmd.NeedsOrg {
					cliCommand.Flags = append(sub.Flags,
						cli.StringFlag{
							Name:  "org",
							Usage: "full or partial organization name.",
						})
				}

				cliCommand.Subcommands[idx].Flags = sub.Flags
			}
		}

		cmds = append(cmds, cliCommand)
	}

	app.Commands = cmds
	app.Name = "cx"
	app.Usage = "Cloud 66 Command line toolbelt (Detailed help: http://help.cloud66.com/toolbelt/toolbelt-introduction)"
	app.Author = "Cloud 66"
	app.Email = "support@cloud66.com"
	app.Version = VERSION
	app.CommandNotFound = suggest
	app.Before = beforeCommand
	app.Action = doMain

	setGlobals(app)
	app.Run(os.Args)
}

func beforeCommand(c *cli.Context) error {
	account := ""
	environment := ""
	if c.GlobalString("account") != "" {
		account = "_" + c.GlobalString("account")
		// remove all cached ssh keys upon account switch
		err := clearSshKeyCache()
		if err != nil {
			return err
		}
	}

	// set the env vars from global options
	if c.GlobalString("runenv") != "production" {
		environment = "_" + c.GlobalString("runenv")

		fmt.Printf(ansi.Color(fmt.Sprintf("Running against %s environment\n", c.GlobalString("runenv")), "grey"))
	}

	if account != "" || environment != "" {
		tokenFile = "cx" + account + environment + ".json"
	}

	if c.GlobalString("fayeEndpoint") != "" {
		fayeEndpoint = c.GlobalString("fayeEndpoint")
	}

	debugMode = c.GlobalBool("debug")

	var command string
	if len(c.Args()) >= 1 {
		command = c.Args().First()
	}

	clientId = os.Getenv("CX_APP_ID")
	clientSecret = os.Getenv("CX_APP_SECRET")

	if clientId == "" {
		clientId = "d4631fd51633bef0c04c6f946428a61fb9089abf4c1e13c15e9742cafd84a91f"
	}
	if clientSecret == "" {
		clientSecret = "e663473f7b991504eb561e208995de15550f499b6840299df588cebe981ba48e"
	}

	if (command != "version") && (command != "help") && (command != "update") && (command != "test") {
		initClients(c, true)
	}

	if command == "test" {
		initClients(c, false)
	}

	if (command != "update") && (VERSION != "dev") {
		defer backgroundRun()
	}

	return nil
}

func setGlobals(app *cli.App) {
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "runenv",
			Usage:  "sets the environment this toolbelt is running against",
			Value:  "production",
			EnvVar: "CXENVIRONMENT",
		},
		cli.StringFlag{
			Name:  "account",
			Usage: "switches between different Cloud 66 profiles (this is a cx client profile)",
			Value: "",
		},
		cli.StringFlag{
			Name:  "org",
			Usage: "targets a specific organisation for a command (this is a Cloud 66 Organisation)",
			Value: "",
		},
		cli.StringFlag{
			Name:   "fayeEndpoint",
			Usage:  "sets the Faye endpoint this toolbelt is running against",
			EnvVar: "CX_FAYE_ENDPOINT",
		},
		cli.BoolFlag{
			Name:   "debug",
			Usage:  "run in debug mode",
			EnvVar: "CXDEBUG",
		},
	}
}

func buildBasicCommand() cli.Command {
	return cli.Command{}
}

func doMain(c *cli.Context) {
	cli.ShowAppHelp(c)
}

func initClients(c *cli.Context, startAuth bool) {
	// is there a token file?
	_, err := os.Stat(filepath.Join(cxHome(), tokenFile))
	if err != nil {
		fmt.Println("No previous authentication found.")
		if startAuth {
			cloud66.Authorize(cxHome(), tokenFile, clientId, clientSecret, redirectURL, scope)
			os.Exit(1)
		} else {
			os.Exit(1)
		}
	} else {
		client = cloud66.GetClient(cxHome(), tokenFile, VERSION, "cx", clientId, clientSecret, redirectURL, scope)
		organization, err := org(c)
		if err != nil {
			printFatal("Unable to retrieve organization")
			os.Exit(2)
		}
		if organization != nil {
			client.AccountId = &organization.Id
		}
		debugMode = c.GlobalBool("debug")
		client.Debug = debugMode
	}
}

func recoverPanic() {
	if VERSION != "dev" {
		raven.CapturePanicAndWait(func() {
			if rec := recover(); rec != nil {
				panic(rec)
			}
		}, map[string]string{
			"Version":      VERSION,
			"Platform":     runtime.GOOS,
			"Architecture": runtime.GOARCH,
			"goversion":    runtime.Version()})
	}
}

func filterByEnvironmentExact(item interface{}) bool {
	if flagEnvironment == "" {
		return true
	}
	return strings.ToLower(item.(cloud66.Stack).Environment) == strings.ToLower(flagEnvironment)
}

func filterByEnvironmentFuzzy(item interface{}) bool {
	if flagEnvironment == "" {
		return true
	}
	return strings.HasPrefix(strings.ToLower(item.(cloud66.Stack).Environment), strings.ToLower(flagEnvironment))
}

func org(c *cli.Context) (*cloud66.Account, error) {
	if flagOrg != nil {
		return flagOrg, nil
	}

	if c.String("org") != "" {
		orgs, err := client.AccountInfos()
		if err != nil {
			return nil, err
		}

		var orgNames []string
		for _, org := range orgs {
			if org.Name == "" {
				return nil, errors.New("One or more of the organizations you are a member of don't have a name. Please make sure you name the organizations.")
			}
			orgNames = append(orgNames, org.Name)
		}
		idx, err := fuzzyFind(orgNames, c.String("org"), false)
		if err != nil {
			return nil, err
		}

		flagOrg = &orgs[idx]
	} else {
		flagOrg = nil
	}

	return flagOrg, nil
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
		stacks, err := client.StackListWithFilter(filterByEnvironmentExact)
		if err != nil {
			return nil, err
		}
		var stackNames []string
		for _, stack := range stacks {
			stackNames = append(stackNames, stack.Name)
		}
		idx, err := fuzzyFind(stackNames, c.String("stack"), false)
		if err != nil {
			// try fuzzy env match
			stacks, err = client.StackListWithFilter(filterByEnvironmentFuzzy)
			if err != nil {
				return nil, err
			}
			var stackFuzzNames []string
			for _, stack := range stacks {
				stackFuzzNames = append(stackFuzzNames, stack.Name)
			}
			idx, err = fuzzyFind(stackFuzzNames, c.String("stack"), false)
			if err != nil {
				return nil, err
			}
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

func mustServer(c *cli.Context, stack cloud66.Stack, flagServer string, ignoreDocker bool) *cloud66.Server {
	servers, err := client.Servers(stack.Uid)
	if err != nil {
		printFatal(err.Error())
	}
	server, err := findServer(servers, flagServer)
	if err != nil {
		printFatal(err.Error())
	}
	if server == nil {
		printFatal("Server '" + flagServer + "' not found")
	}
	if !ignoreDocker && !server.HasRole("docker") && !server.HasRole("kubes") {
		printFatal("Server '" + flagServer + "' is not a docker server")
	}
	fmt.Printf("Server: %s\n", server.Name)
	return server
}

func mustOrg(c *cli.Context) *cloud66.Account {
	org, err := org(c)
	if err != nil {
		printFatal(err.Error())
	}

	if org == nil {
		printFatal("No organization specified. Please use --org flag")
	}

	return org
}

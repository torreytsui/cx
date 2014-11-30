package main

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"text/tabwriter"

	"github.com/cloud66/cloud66"
)

var flagUnmanaged bool

var cmdInfo = &Command{
	Run:      runInfo,
	Usage:    "info [-s <stack>] [-e <environment>] [-unmanaged]",
	Category: "cx",
	Short:    "shows information about your account, toolbelt and the current directory or the specified stack",
	Long: `info lists the account information, toolbelt information and if applicable information about the
  your current directory.
  Use unmanaged parameter to list the servers under your cloud account that are NOT in any of your stacks.`,
}

func init() {
	cmdInfo.Flag.StringVar(&flagStackName, "s", "", "stack name")
	cmdInfo.Flag.StringVar(&flagEnvironment, "e", "", "stack environment")
	cmdInfo.Flag.BoolVar(&flagUnmanaged, "unmanaged", false, "list unmanaged servers")
}

func runInfo(cmd *Command, args []string) {
	if err := toolbeltInfo(); err != nil {
		printFatal(err.Error())
	}
	if err := accountInfo(); err != nil {
		printFatal(err.Error())
	}
	if err := stackInfo(); err != nil {
		printFatal(err.Error())
	}
}

func accountInfo() error {
	accountInfos, err := client.AccountInfos()
	if err != nil {
		return err
	}

	if len(accountInfos) < 1 {
		printFatal("User associated with this request returning zero references")
		os.Exit(2)
	}

	var currentAccountId int

	for _, accountInfo := range accountInfos {
		fmt.Printf("\n")
		fmt.Printf("Account owner: %s\n", accountInfo.Owner)
		fmt.Printf("Running %d stack(s)\n", accountInfo.StackCount)
		fmt.Printf("Used clouds: %s\n", strings.Join(accountInfo.UsedClouds, ", "))

		if accountInfo.CurrentAccount {
			currentAccountId = accountInfo.Id
		}
	}

	if flagUnmanaged {
		w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
		defer w.Flush()

		fmt.Println("\nFetching the list of unmanaged servers...")
		mainAccount, err := client.AccountInfo(currentAccountId, true)
		if err != nil {
			return err
		}

		printUnmanagedServerList(w, mainAccount.UnmanagedServers)
	}

	return nil
}

func stackInfo() error {
	stack, err := stack()
	if err != nil {
		return err
	}

	if stack != nil {
		fmt.Println()
		fmt.Printf("Stack info: %s (%s)\n", stack.Name, stack.Environment)
		fmt.Printf("Uid: %s\n", stack.Uid)
		fmt.Printf("Git: %s (%s)\n", stack.Git, stack.GitBranch)
		fmt.Printf("Hosted on: %s\n", stack.Cloud)
		fmt.Printf("FQDN: %s\n", stack.Fqdn)
		fmt.Printf("Framework: %s (%s)\n", stack.Framework, stack.Language)
	}
	return nil
}

func toolbeltInfo() error {
	fmt.Println("Cloud 66 Toolbelt (c) 2014 Cloud66 Ltd.")
	if VERSION == "dev" {
		fmt.Println("Development version")
	} else {
		fmt.Printf("v%s\n", VERSION)
	}
	if debugMode {
		fmt.Println("Running in Debug mode")
	}
	fmt.Printf("OS: %s, Architecture: %s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Println("For more information visit http://cloud-66-help.c66.me/introduction-to-cloud-66/introduction-to-cloud-66")

	return nil
}

func printUnmanagedServerList(w io.Writer, servers []cloud66.UnmanagedServer) {
	for _, a := range servers {
		listUnmanagedServer(w, a)
	}
}

func listUnmanagedServer(w io.Writer, a cloud66.UnmanagedServer) {
	listRec(w,
		a.Vendor,
		a.Id,
	)
}

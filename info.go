package main

import (
	"fmt"
	"runtime"
	"strings"
)

var cmdInfo = &Command{
	Run:      runInfo,
	Usage:    "info",
	Category: "cx",
	Short:    "shows information about your account, toolbelt and the current directory",
	Long: `info lists the account information, toolbelt information and if applicable information about the
  your current directory.`,
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
	userInfo, err := client.UserInfo()
	if err != nil {
		return err
	}

	fmt.Printf("Account owner: %s\n", userInfo.Owner)
	fmt.Printf("Running %d stack(s)\n", userInfo.StackCount)
	fmt.Printf("Used clouds: %s\n", strings.Join(userInfo.UsedClouds, ", "))
	return nil
}

func stackInfo() error {
	stack, err := stack()
	if err != nil {
		return err
	}

	if stack != nil {
		fmt.Printf("Current stack: %s\n", stack.Name)
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
	fmt.Println("For more information visit http://help.cloud66.com/cloud-66-toolbelt/introduction.html")

	return nil
}

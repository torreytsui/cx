package main

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/cloud66/cx/cloud66"
)

var _ = fmt.Print
var _ = os.Stdout

func areSameRemotes(lhs string, rhs string) (bool, error) {
	lhs, rhs = strings.TrimSpace(lhs), strings.TrimSpace(rhs)

	if debugMode {
		fmt.Printf("Comparing '%s' and '%s'\n", lhs, rhs)
	}
	// are they the same?
	if lhs == rhs {
		return true, nil
	}

	// if the different is only the .git at the end, then they are the same
	if lhs + ".git" == rhs || rhs + ".git" == lhs {
		return true, nil
	}

	if strings.HasPrefix(lhs, "git@") {
		lhs = strings.Replace(lhs, ":", "/", 1)
		lhs = strings.Replace(lhs, "git@", "git://", 1)
	}

	if strings.HasPrefix(rhs, "git@") {
		rhs = strings.Replace(rhs, ":", "/", 1)
		rhs = strings.Replace(rhs, "git@", "git://", 1)
	}

	lhsParsed, err := url.Parse(lhs)
	if err != nil {
		return false, err
	}

	rhsParsed, err := url.Parse(rhs)
	if err != nil {
		return false, err
	}

	// http and https are the same
	if rhsParsed.Path == lhsParsed.Path {
		return true, nil
	}

	return false, nil
}

func localGitBranch() string {
	b, err := exec.Command("git", "name-rev", "--name-only", "HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func remoteGitUrl() string {
	b, err := exec.Command("git", "config", "remote.origin.url").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

// search for the given stack using the git URL and branch
func stackFromGitRemote(gitUrl string, gitBranch string) (*cloud66.Stack, error) {
	stacks, err := client.StackList()
	for _, stack := range stacks {
		r, err := areSameRemotes(stack.Git, gitUrl)
		if err != nil {
			return nil, err
		}
		if r && stack.GitBranch == gitBranch {
			return &stack, nil
		}
	}

	return nil, err
}

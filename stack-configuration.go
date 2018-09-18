package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	"github.com/cloud66-oss/cloud66"
	"github.com/cloud66/cli"
)

type ConfigurationFile struct {
	Name         string
	Type         string
	ListFunc     func(*cli.Context, string)
	DownloadFunc func(*cli.Context, string)
	UploadFunc   func(*cli.Context, string)
}

//-------------------------------------------------------------------------------------

func runStackConfigurationList(c *cli.Context) {
	stack := mustStack(c)
	configurations, err := client.ConfigurationList(stack.Uid)
	must(err)
	printConfigurationList(configurations)
}

func printConfigurationList(configurations []cloud66.Configuration) {
	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()
	listRec(w,
		"TYPE",
		"FILE NAME",
		"BODY SIZE",
		"UPDATED",
		"CHANGED BY",
		"COMMENTS",
	)
	for _, configuration := range configurations {
		listRec(w,
			configuration.Type,
			configuration.Name,
			len(configuration.Body),
			prettyTime{configuration.UpdatedAt},
			configuration.ChangedBy,
			configuration.Comments,
		)
	}
}

//-------------------------------------------------------------------------------------

func runStackConfigurationDownload(c *cli.Context) {
	var err error
	theType := c.String("type")
	if theType == "" {
		printFatal("No type specified")
	}
	output := c.String("output")
	if output != "" {
		// convert to abs path
		output, err = filepath.Abs(output)
		must(err)
	}
	stack := mustStack(c)
	configuration, err := client.ConfigurationDownload(stack.Uid, theType)
	must(err)
	if output == "" {
		// output configuration body to STDOUT
		printConfiguration(*configuration)
	} else {
		// save configuration body to output file
		saveConfiguration(*configuration, output)
	}
}

func printConfiguration(configuration cloud66.Configuration) {
	fmt.Println(configuration.Body)
}

func saveConfiguration(configuration cloud66.Configuration, output string) {
	bodyBytes := []byte(configuration.Body)
	err := ioutil.WriteFile(output, bodyBytes, 0644)
	must(err)
	if len(configuration.Body) > 0 {
		fmt.Println("Configuration saved to file successfully")
	} else {
		fmt.Println("Configuration saved to file successfully (NOTE: the configuration is empty!)")
	}
}

//-------------------------------------------------------------------------------------

func runStackConfigurationUpload(c *cli.Context) {
	var err error
	theType := c.String("type")
	if theType == "" {
		printFatal("No type specified")
	}
	source := c.String("source")
	if source == "" {
		printFatal("No source specified")
	}
	noApply := c.Bool("no-apply")
	commitMessage := c.String("commit-message")

	source, err = filepath.Abs(source)
	must(err)
	bodyBytes, err := ioutil.ReadFile(source)
	must(err)
	// convert source file into string contents
	body := string(bodyBytes)
	// load validate the stack
	stack := mustStack(c)
	// ensure this configuration exists
	configuration, err := client.ConfigurationDownload(stack.Uid, theType)
	must(err)
	// set the new configuration body
	mustApply := !noApply

	asyncRes, err := client.ConfigurationUpload(stack.Uid, theType, commitMessage, body, mustApply)
	must(err)

	genericRes, err := client.WaitStackAsyncAction(asyncRes.Id, stack.Uid, 5*time.Second, 20*time.Minute, true)
	must(err)

	var successMessage string
	if !configuration.CanApply {
		if mustApply {
			successMessage = "Configuration uploaded (apply not supported for given configuration)"
		} else {
			successMessage = "Configuration uploaded"
		}
	} else {
		if mustApply {
			successMessage = "Configuration uploaded and applied!"
		} else {
			successMessage = "Configuration uploaded (not applied)"
		}
	}
	printGenericResponseCustom(*genericRes, successMessage, "")
}

//-------------------------------------------------------------------------------------

func runStackConfigurationApply(c *cli.Context) {
	var err error
	theType := c.String("type")
	if theType == "" {
		printFatal("No type specified")
	}

	// load validate the stack
	stack := mustStack(c)
	// ensure this configuration exists
	configuration, err := client.ConfigurationDownload(stack.Uid, theType)
	must(err)

	if !configuration.CanApply {
		printFatal("Apply is not supported for the given configuration type")
	}

	asyncRes, err := client.ConfigurationApply(stack.Uid, theType)
	must(err)
	genericRes, err := client.WaitStackAsyncAction(asyncRes.Id, stack.Uid, 5*time.Second, 20*time.Minute, true)
	must(err)
	successMessage := "Configuration applied"
	printGenericResponseCustom(*genericRes, successMessage, "")
}

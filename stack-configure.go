package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"text/tabwriter"

	"github.com/cloud66/cloud66"
	"github.com/cloud66/cx/term"

	"github.com/cloud66/cli"
)

type ConfigureFile struct {
	Name         string
	ListFunc     func(*cli.Context, string)
	DownloadFunc func(*cli.Context, string)
	UploadFunc   func(*cli.Context, string)
}

func mustFile(c *cli.Context) ConfigureFile {
	var file ConfigureFile
	switch c.String("file") {
	case "service.yml":
		file = ConfigureFile{
			Name:         c.String("file"),
			ListFunc:     runServiceYamlList,
			DownloadFunc: runDownloadServiceYaml,
			UploadFunc:   runUploadServiceYaml}
	case "manifest.yml":
		file = ConfigureFile{
			Name:         c.String("file"),
			ListFunc:     runManifestYamlList,
			DownloadFunc: runDownloadManifestYaml,
			UploadFunc:   runUploadManifestYaml}
	default:
		printFatal("No file type specified or file type is not supported. Use --file flag to choose file type. supported values are: service.yml , manifest.yml")
	}

	return file
}

func runStackConfigureFileList(c *cli.Context) {
	stack := mustStack(c)
	file := mustFile(c)

	file.ListFunc(c, stack.Uid)
}

func runStackConfigureDownloadFile(c *cli.Context) {
	stack := mustStack(c)
	file := mustFile(c)

	file.DownloadFunc(c, stack.Uid)
}

func runStackConfigureUploadFile(c *cli.Context) {
	stack := mustStack(c)
	file := mustFile(c)

	file.UploadFunc(c, stack.Uid)
}

// --
func runServiceYamlList(c *cli.Context, stackUid string) {
	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()

	serviceYamls, err := client.ServiceYamlList(stackUid, false)
	must(err)

	printServiceYamlList(w, serviceYamls)
}

func printServiceYamlList(w io.Writer, serviceYamls []cloud66.ServiceYaml) {
	for _, a := range serviceYamls {
		if a.Uid != "" {
			listRec(w,
				a.Uid,
				a.Comments,
				prettyTime{a.CreatedAt},
			)
		}
	}
}

func runManifestYamlList(c *cli.Context, stackUid string) {
	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()

	manifestYamls, err := client.ManifestYamlList(stackUid, false)
	must(err)

	printManifestYamlList(w, manifestYamls)
}

func printManifestYamlList(w io.Writer, manifestYamls []cloud66.ManifestYaml) {
	for _, a := range manifestYamls {
		if a.Uid != "" {
			listRec(w,
				a.Uid,
				a.Comments,
				prettyTime{a.CreatedAt},
			)
		}
	}
}

func runDownloadServiceYaml(c *cli.Context, stackUid string) {
	version := c.String("version")
	if version == "" {
		version = "latest"
	}

	serviceYaml, err := client.ServiceYamlInfo(stackUid, version)
	must(err)

	output := c.String("output")
	if output != "" {
		err := writeFile(output, serviceYaml.Body)
		must(err)
	} else {
		fmt.Println(serviceYaml.Body)
	}
}

func runDownloadManifestYaml(c *cli.Context, stackUid string) {
	version := c.String("version")
	if version == "" {
		version = "latest"
	}

	manifestYaml, err := client.ManifestYamlInfo(stackUid, version)
	must(err)

	output := c.String("output")
	if output != "" {
		err := writeFile(output, manifestYaml.Body)
		must(err)
	} else {
		fmt.Println(manifestYaml.Body)
	}
}

func runUploadServiceYaml(c *cli.Context, stackUid string) {
	serviceYamlFile := c.Args().First()
	if serviceYamlFile == "" {
		printFatal("service_yaml file path is required")
	} else {
		serviceYamlFile = expandPath(serviceYamlFile)
	}
	serviceYamlBytes, err := ioutil.ReadFile(serviceYamlFile)
	must(err)
	serviceYaml := string(serviceYamlBytes)

	comments := c.String("comments")
	if comments == "" {
		fmt.Println("\nComments can't be blank, Please add one:")
		if term.IsTerminal(os.Stdin) {
			fmt.Printf("> ")
		}

		reader := bufio.NewReader(os.Stdin)
		if comments, err = reader.ReadString('\n'); err != nil {
			printFatal(err.Error())
		}
	}

	_, err = client.CreateServiceYaml(stackUid, serviceYaml, comments)
	must(err)
}

func runUploadManifestYaml(c *cli.Context, stackUid string) {
	manifestYamlFile := c.Args().First()
	if manifestYamlFile == "" {
		printFatal("manifest_yaml file path is required")
	} else {
		manifestYamlFile = expandPath(manifestYamlFile)
	}
	manifestYamlBytes, err := ioutil.ReadFile(manifestYamlFile)
	must(err)
	manifestYaml := string(manifestYamlBytes)

	comments := c.String("comments")
	if comments == "" {
		fmt.Println("\nComments can't be blank, Please add one:")
		if term.IsTerminal(os.Stdin) {
			fmt.Printf("> ")
		}

		reader := bufio.NewReader(os.Stdin)
		if comments, err = reader.ReadString('\n'); err != nil {
			printFatal(err.Error())
		}
	}

	_, err = client.CreateManifestYaml(stackUid, manifestYaml, comments)
	must(err)
}

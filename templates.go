package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/cloud66-oss/cloud66"
	"github.com/cloud66/cli"
)

var cmdTemplates = &Command{
	Name:       "templates",
	Build:      buildTemplates,
	NeedsStack: false,
	NeedsOrg:   true,
	Short:      "stencil template repository management",
}

func buildTemplates() cli.Command {
	base := buildBasicCommand()
	base.Subcommands = []cli.Command{
		cli.Command{
			Name:   "list",
			Usage:  "shows the list of all stencil template repositories in an account",
			Action: runListTemplates,
			Description: `

Examples:
$ cx --org='My Awesome Organization' templates list
Name                          Uid                                  Git Repository                                           Git Branch  Status
First Awesome Repository      bt-2e0810a17c33ab35d7970ff330b1f916  git@github.com:AwesomeOrganization/awesome-stencils.git  test        Available
Second Awesome Repository     bt-e2e869ee6ce97ee58a17aa264bed1e0c  git@github.com:AwesomeOrganization/better-stencils.git   test        Available
`,
		},
		cli.Command{
			Name:  "resync",
			Usage: "pulls the latest code from the stencil template repository",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "template,t",
					Usage: "template UID",
				},
			},
			Action: runResyncTemplate,
			Description: `

Examples:
$ cx --org='My Awesome Organization' templates resync 'First Awesome Repository'
`,
		},
	}

	return base
}

func runListTemplates(c *cli.Context) {
	mustOrg(c)

	baseTemplates, err := client.ListBaseTemplates()
	if err != nil {
		printFatal(err.Error())
	}

	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()

	printBaseTemplates(w, baseTemplates)
}

func runResyncTemplate(c *cli.Context) {
	mustOrg(c)

	baseTemplateUID := c.String("template")
	if baseTemplateUID == "" {
		printFatal("No template UID specified. Please use the --template flag to specify one.")
	}

	baseTemplates, err := client.ListBaseTemplates()
	if err != nil {
		printFatal(err.Error())
	}

	requestedBaseTemplateIndex, err := getBaseTemplateIndexByUID(baseTemplates, baseTemplateUID)
	if err != nil {
		printFatal(err.Error())
	}

	baseTemplate, err := client.SyncBaseTemplate(baseTemplates[requestedBaseTemplateIndex].Uid)
	if err != nil {
		printFatal(err.Error())
	}

	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()

	printBaseTemplate(w, baseTemplate)
}

func printBaseTemplates(w io.Writer, baseTemplates []cloud66.BaseTemplate) {
	printBaseTemplateHeader(w)
	for _, baseTemplate := range baseTemplates {
		printBaseTemplateRow(w, &baseTemplate)
	}
}

func printBaseTemplate(w io.Writer, baseTemplate *cloud66.BaseTemplate) {
	printBaseTemplateHeader(w)
	printBaseTemplateRow(w, baseTemplate)
}

func printBaseTemplateHeader(w io.Writer) {
	listRec(w,
		"Name",
		"Uid",
		"Git Repository",
		"Git Branch",
		"Status",
	)
}

func printBaseTemplateRow(w io.Writer, baseTemplate *cloud66.BaseTemplate) {
	listRec(w,
		baseTemplate.Name,
		baseTemplate.Uid,
		baseTemplate.GitRepo,
		baseTemplate.GitBranch,
		baseTemplate.Status(),
	)
}

func getBaseTemplateIndexByUID(baseTemplates []cloud66.BaseTemplate, baseTemplateUID string) (int, error) {
	for i, baseTemplate := range baseTemplates {
		if baseTemplate.Uid == baseTemplateUID {
			return i, nil
		}
	}

	return -1, errors.New(fmt.Sprintf("Could not find template repository with UID %s.", baseTemplateUID))
}

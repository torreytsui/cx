package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"text/tabwriter"

	"github.com/cloud66-oss/cloud66"
	"github.com/cloud66/cli"
)

var cmdFormations = &Command{
	Name:       "formations",
	Build:      buildFormations,
	Short:      "commands to work with formations",
	NeedsStack: true,
	NeedsOrg:   false,
}

func buildFormations() cli.Command {
	base := buildBasicCommand()
	base.Subcommands = []cli.Command{
		cli.Command{
			Name:   "list",
			Action: runFormations,
			Usage:  "lists all the formations of a stack.",
			Description: `List all the formations of a stack.
The information contains the name and UUID

Examples:
$ cx formations list -s mystack
$ cx formations list -s mystack foo bar // only show formations foo and bar
`,
		},
		cli.Command{
			Name:  "bundle",
			Usage: "formation bundle commands",
			Subcommands: []cli.Command{
				cli.Command{
					Name:   "download",
					Action: runDownloadBundle,
					Usage:  "Specify the formation to use",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "formation",
							Usage: "Specify the formation to use",
						},
						cli.StringFlag{
							Name:  "stack,s",
							Usage: "full or partial stack name. This can be omitted if the current directory is a stack directory",
						},
						cli.StringFlag{
							Name:  "file",
							Usage: "filename for the bundle file. formation extension will be added",
						},
						cli.BoolFlag{
							Name:  "overwrite",
							Usage: "overwrite existing bundle file is it exists",
						},
					},
				},
			},
		},
		cli.Command{
			Name:  "stencils",
			Usage: "formation stencil commands",
			Subcommands: []cli.Command{
				cli.Command{
					Name:   "list",
					Usage:  "List all formation stencils",
					Action: runStencilsList,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "formation",
							Usage: "Specify the formation to use",
						},
						cli.StringFlag{
							Name:  "stack,s",
							Usage: "full or partial stack name. This can be omitted if the current directory is a stack directory",
						},
						cli.StringFlag{
							Name:  "output,o",
							Usage: "tailor output view (standard|wide)",
						},
					},
					Description: `Fetch all formation stencils and their templates
Examples:
$ cx formations stencils list --formation foo
$ cx formations stencils list --formation bar
`,
				},
				cli.Command{
					Name:   "show",
					Usage:  "Shows the content of a single stencil",
					Action: runShowStencil,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "formation",
							Usage: "Specify the formation to use",
						},
						cli.StringFlag{
							Name:  "stack,s",
							Usage: "full or partial stack name. This can be omitted if the current directory is a stack directory",
						},
						cli.StringFlag{
							Name:  "stencil",
							Usage: "Stencil filename",
						},
					},
				},
			},
		},
	}

	return base
}

func runFormations(c *cli.Context) {
	stack := mustStack(c)
	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()

	var formations []cloud66.Formation
	var err error
	formations, err = client.Formations(stack.Uid, false)
	must(err)

	formationNames := c.Args()

	for idx, i := range formationNames {
		formationNames[idx] = strings.ToLower(i)
	}
	sort.Strings(formationNames)
	if len(formationNames) == 0 {
		printFormationList(w, formations)
	} else {
		// filter out the unwanted formations
		var filteredFormations []cloud66.Formation
		for _, i := range formations {
			sorted := sort.SearchStrings(formationNames, strings.ToLower(i.Name))
			if sorted < len(formationNames) && strings.ToLower(formationNames[sorted]) == strings.ToLower(i.Name) {
				filteredFormations = append(filteredFormations, i)
			}
		}
		printFormationList(w, filteredFormations)
	}
}

func runStencilsList(c *cli.Context) {
	stack := mustStack(c)
	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()

	formationName := c.String("formation")
	if formationName == "" {
		printFatal("No formation provided. Please use --formation to specify a formation")
	}

	var formations []cloud66.Formation
	var err error
	formations, err = client.Formations(stack.Uid, true)
	must(err)

	output := c.String("output")
	if output == "" {
		output = "standard"
	}

	for _, formation := range formations {
		if formation.Name == formationName {
			printStencils(w, formation, output)
			return
		}
	}

	printFatal("No formation named '%s' found", formationName)
}

func runShowStencil(c *cli.Context) {
	stack := mustStack(c)
	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()

	formationName := c.String("formation")
	if formationName == "" {
		printFatal("No formation provided. Please use --formation to specify a formation")
	}

	stencilName := c.String("stencil")
	if stencilName == "" {
		printFatal("No stencil name provided. Please use --stencil to specify a formation")
	}

	var formations []cloud66.Formation
	var err error
	formations, err = client.Formations(stack.Uid, true)
	must(err)

	foundStencil := false

	for _, formation := range formations {
		if formation.Name == formationName {
			for _, stencil := range formation.Stencils {
				if stencil.Filename == stencilName {
					printStencil(stencil)
					foundStencil = true
				}
			}

			if !foundStencil {
				printFatal("No stencil named '%s' found", stencilName)
			}
			return
		}
	}

	printFatal("No formation named '%s' found", formationName)
}

func runDownloadBundle(c *cli.Context) {
	stack := mustStack(c)

	formationName := c.String("formation")
	if formationName == "" {
		printFatal("No formation provided. Please use --formation to specify a formation")
	}

	bundleFile := c.String("file")
	if bundleFile == "" {
		bundleFile = formationName
	}

	if filepath.Ext(bundleFile) != ".formation" {
		bundleFile = bundleFile + ".formation"
	}

	if _, err := os.Stat(bundleFile); err == nil {
		if !c.Bool("overwrite") {
			printFatal("%s already exists", bundleFile)
		}
	}

	fmt.Println("Fetching bundle from the server...")
	var formations []cloud66.Formation
	var err error
	formations, err = client.Formations(stack.Uid, true)
	must(err)

	for _, formation := range formations {
		if formation.Name == formationName {
			bundleFormation(formation, bundleFile)
			return
		}
	}

	printFatal("No formation named '%s' found", formationName)
}

func printStencils(w io.Writer, formation cloud66.Formation, output string) {
	stencils := formation.Stencils
	sort.Sort(stencilBySequence(stencils))

	if output == "standard" {
		listRec(w,
			"UID",
			"FILENAME",
			"TAGS",
			"CREATED AT",
			"LAST UPDATED")
	} else {
		listRec(w,
			"UID",
			"FILENAME",
			"SERVICE",
			"TAGS",
			"TEMPLATE",
			"GITFILE",
			"INLINE",
			"CREATED AT",
			"LAST UPDATED")
	}

	for _, a := range stencils {
		listStencil(w, a, output)
	}
}

func bundleFormation(formation cloud66.Formation, bundleFile string) {
	// build a temp folder structure
	dir, err := ioutil.TempDir("", fmt.Sprintf("%s-formation-bundle-", formation.Name))
	if err != nil {
		printFatal(err.Error())
	}

	//defer os.RemoveAll(dir)
	stencilsDir := filepath.Join(dir, "stencils")
	err = os.Mkdir(stencilsDir, os.ModePerm)
	if err != nil {
		printFatal(err.Error())
	}
	stencilGroupsDir := filepath.Join(dir, "stencil-groups")
	err = os.Mkdir(stencilGroupsDir, os.ModePerm)
	if err != nil {
		printFatal(err.Error())
	}
	policiesDir := filepath.Join(dir, "policies")
	err = os.Mkdir(policiesDir, os.ModePerm)
	if err != nil {
		printFatal(err.Error())
	}
	manifestFilename := filepath.Join(dir, "manifest.json")

	// save all the files individually
	// stencils
	fmt.Println("Saving stencils...")
	for _, stencil := range formation.Stencils {
		fileName := filepath.Join(stencilsDir, stencil.Uid+".yml")
		file, err := os.Create(fileName)
		defer file.Close()
		if err != nil {
			printFatal(err.Error())
		}

		file.WriteString(stencil.Body)
	}

	// stencilgroups
	fmt.Println("Saving stencil groups...")
	for _, stencilGroup := range formation.StencilGroups {
		fileName := filepath.Join(stencilGroupsDir, stencilGroup.Uid+".json")
		file, err := os.Create(fileName)
		defer file.Close()
		if err != nil {
			printFatal(err.Error())
		}

		file.WriteString(stencilGroup.Rules)
	}

	// policies
	fmt.Println("Saving policies...")
	for _, policy := range formation.Policies {
		fileName := filepath.Join(policiesDir, policy.Uid+".cop")
		file, err := os.Create(fileName)
		defer file.Close()
		if err != nil {
			printFatal(err.Error())
		}

		file.WriteString(policy.Body)
	}

	// create and save the manifest
	fmt.Println("Saving bundle manifest...")
	manifest := cloud66.CreateFormationBundle(formation, fmt.Sprintf("cx (%s)", VERSION))
	buf, err := json.MarshalIndent(manifest, "", "    ")
	if err != nil {
		printFatal(err.Error())
	}
	manifestFile, err := os.Create(manifestFilename)
	if err != nil {
		printFatal(err.Error())
	}
	defer manifestFile.Close()

	_, err = manifestFile.Write(buf)
	if err != nil {
		printFatal(err.Error())
	}

	// tarball
	err = Tar(dir, bundleFile)
	if err != nil {
		printFatal(err.Error())
	}
	fmt.Printf("Bundle is saved to %s\n", bundleFile)
}

func printStencil(stencil cloud66.Stencil) {
	var buffer bytes.Buffer

	buffer.WriteString(stencil.Body)
	fmt.Print(buffer.String())
}

func printFormationList(w io.Writer, formations []cloud66.Formation) {
	sort.Sort(formationByName(formations))

	listRec(w,
		"UID",
		"NAME",
		"TAGS",
		"STENCILS",
		"STENCIL GROUPS",
		"POLICIES",
		"BASE TEMPLATE",
		"CREATED AT",
		"LAST UPDATED")

	for _, a := range formations {
		if a.Name != "" {
			listFormation(w, a)
		}
	}
}

func listFormation(w io.Writer, a cloud66.Formation) {
	ta := a.CreatedAt

	listRec(w,
		a.Uid,
		a.Name,
		a.Tags,
		len(a.Stencils),
		len(a.StencilGroups),
		len(a.Policies),
		a.BaseTemplate,
		prettyTime{ta},
		prettyTime{a.UpdatedAt},
	)
}

func listStencil(w io.Writer, a cloud66.Stencil, output string) {
	ta := a.CreatedAt

	if output == "standard" {
		listRec(w,
			a.Uid,
			a.Filename,
			a.Tags,
			prettyTime{ta},
			prettyTime{a.UpdatedAt},
		)
	} else {
		listRec(w,
			a.Uid,
			a.Filename,
			a.ContextID,
			a.Tags,
			a.TemplateFilename,
			a.GitfilePath,
			a.Inline,
			prettyTime{ta},
			prettyTime{a.UpdatedAt})
	}
}

type formationByName []cloud66.Formation

func (a formationByName) Len() int           { return len(a) }
func (a formationByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a formationByName) Less(i, j int) bool { return a[i].Name < a[j].Name }

type stencilBySequence []cloud66.Stencil

func (a stencilBySequence) Len() int           { return len(a) }
func (a stencilBySequence) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a stencilBySequence) Less(i, j int) bool { return a[i].Sequence < a[j].Sequence }

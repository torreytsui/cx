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
			Action: runListFormations,
			Usage:  "lists all the formations of a stack.",
			Description: `List all the formations of a stack.
The information contains the name and UUID

Examples:
$ cx formations list -s mystack
$ cx formations list -s mystack foo bar // only show formations foo and bar
`,
		},
		cli.Command{
			Name:   "create",
			Action: runCreateFormation,
			Usage:  "Create a new formation",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "name",
					Usage: "Formation name",
				},
				cli.StringFlag{
					Name:  "template-repo",
					Usage: "Base Template repository URL",
				},
				cli.StringFlag{
					Name:  "template-branch",
					Usage: "Base Template repository branch",
				},
				cli.StringFlag{
					Name:  "tags",
					Usage: "Formation tags",
				},
			},
		},
		cli.Command{
			Name:  "bundle",
			Usage: "formation bundle commands",
			Subcommands: []cli.Command{
				cli.Command{
					Name:   "download",
					Action: runBundleDownload,
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
				cli.Command{
					Name:   "upload",
					Usage:  "Upload a formation bundle to a new formation",
					Action: runBundleUpload,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "formation",
							Usage: "Name for the new formation",
						},
						cli.StringFlag{
							Name:  "stack,s",
							Usage: "full or partial stack name. This can be omitted if the current directory is a stack directory",
						},
						cli.StringFlag{
							Name:  "file",
							Usage: "filename for the bundle file",
						},
						cli.StringFlag{
							Name:  "message",
							Usage: "Commit message",
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
					Action: runListStencils,
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
				cli.Command{
					Name:   "add",
					Usage:  "Add a stencil to the formation",
					Action: runAddStencil,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "formation",
							Usage: "Specify the formaiton to use",
						},
						cli.StringFlag{
							Name:  "stack,s",
							Usage: "Full or partial stack name. This can be omitted if the current directory is a stack directory",
						},
						cli.StringFlag{
							Name:  "stencil",
							Usage: "Stencil file",
						},
						cli.StringFlag{
							Name:  "service",
							Usage: "Service context of the stencil, if applicable",
						},
						cli.StringFlag{
							Name:  "template",
							Usage: "Template filename",
						},
						cli.IntFlag{
							Name:  "sequence",
							Usage: "Stencil sequence",
						},
						cli.StringFlag{
							Name:  "message",
							Usage: "Commit message",
						},
						cli.StringFlag{
							Name:  "tags",
							Usage: "Comma separated tags",
						},
					},
				},
			},
		},
	}

	return base
}

/* Formations */
func runListFormations(c *cli.Context) {
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

func runCreateFormation(c *cli.Context) {
	stack := mustStack(c)

	tags := []string{}
	name := c.String("name")
	templateRepo := c.String("template-repo")
	templateBranch := c.String("template-branch")
	tagList := c.String("tags")
	if tagList != "" {
		tags = strings.Split(tagList, ",")
	}

	_, err := client.CreateFormation(stack.Uid, name, templateRepo, templateBranch, tags)
	if err != nil {
		printFatal(err.Error())
	}

	fmt.Println("Formation created")
}

func runBundleDownload(c *cli.Context) {
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

func runBundleUpload(c *cli.Context) {
	stack := mustStack(c)

	formationName := c.String("formation")
	if formationName == "" {
		printFatal("No formation provided. Please use --formation to specify a formation")
	}

	bundleFile := c.String("file")
	if bundleFile == "" {
		bundleFile = formationName + ".formation"
	}

	tagList := c.String("tags")
	tags := []string{}
	if tagList == "" {
		tags = strings.Split(tagList, ",")
	}

	// untar the bundle
	bundleTopPath, err := ioutil.TempDir("", fmt.Sprintf("%s-formation-bundle-", formationName))
	if err != nil {
		printFatal(err.Error())
	}

	err = Untar(bundleFile, bundleTopPath)
	if err != nil {
		printFatal(err.Error())
	}
	bundlePath := filepath.Join(bundleTopPath, "bundle")
	manifestFile := filepath.Join(bundlePath, "manifest.json")
	message := c.String("message")
	if message == "" {
		printFatal("No message given. Use --message to provide a message for the commit")
	}

	// load the bundle manifest
	bundle, err := os.Open(manifestFile)
	if err != nil {
		printFatal(err.Error())
	}
	defer bundle.Close()

	buff, err := ioutil.ReadAll(bundle)
	if err != nil {
		printFatal(err.Error())
	}

	var fb *cloud66.FormationBundle
	err = json.Unmarshal(buff, &fb)
	if err != nil {
		printFatal(err.Error())
	}

	// create the formation
	fmt.Printf("Creating %s formation...\n", formationName)
	formation, err := client.CreateFormation(stack.Uid, formationName, fb.BaseTemplate.Repo, fb.BaseTemplate.Branch, tags)
	if err != nil {
		printFatal(err.Error())
	}
	fmt.Println("Formation created")

	// add stencils
	fmt.Println("Adding stencils...")
	stencils := make([]*cloud66.Stencil, len(fb.Stencils))
	for idx, stencil := range fb.Stencils {
		stencils[idx], err = stencil.AsStencil(bundlePath)
		if err != nil {
			printFatal(err.Error())
		}
	}
	_, err = client.AddStencils(stack.Uid, formation.Uid, stencils, message)
	if err != nil {
		printFatal(err.Error())
	}
	fmt.Println("Stencils added")

	// TODO: add stencil groups
	// TODO: add policies
}

func bundleFormation(formation cloud66.Formation, bundleFile string) {
	// build a temp folder structure
	topDir, err := ioutil.TempDir("", fmt.Sprintf("%s-formation-bundle-", formation.Name))
	if err != nil {
		printFatal(err.Error())
	}
	dir := filepath.Join(topDir, "bundle")

	//defer os.RemoveAll(dir)
	stencilsDir := filepath.Join(dir, "stencils")
	err = os.MkdirAll(stencilsDir, os.ModePerm)
	if err != nil {
		printFatal(err.Error())
	}
	stencilGroupsDir := filepath.Join(dir, "stencil-groups")
	err = os.MkdirAll(stencilGroupsDir, os.ModePerm)
	if err != nil {
		printFatal(err.Error())
	}
	policiesDir := filepath.Join(dir, "policies")
	err = os.MkdirAll(policiesDir, os.ModePerm)
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

type formationByName []cloud66.Formation

func (a formationByName) Len() int           { return len(a) }
func (a formationByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a formationByName) Less(i, j int) bool { return a[i].Name < a[j].Name }

/* End Formations */

/* Stencils */
func runListStencils(c *cli.Context) {
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

func runAddStencil(c *cli.Context) {
	stack := mustStack(c)

	formationName := c.String("formation")
	if formationName == "" {
		printFatal("No formation provided. Please use --formation to specify a formation")
	}

	stencilFile := c.String("stencil")
	if stencilFile == "" {
		printFatal("No stencil filename provided. Please use --stencil to specify a stencil file")
	}

	tags := []string{}
	contextID := c.String("service")
	template := c.String("template")
	sequence := c.Int("sequence")
	message := c.String("message")
	tagList := c.String("tags")
	if tagList != "" {
		tags = strings.Split(tagList, ",")
	}

	var formations []cloud66.Formation
	var err error
	formations, err = client.Formations(stack.Uid, true)
	must(err)
	var foundFormation cloud66.Formation

	for _, formation := range formations {
		if formation.Name == formationName {
			for _, stencil := range formation.Stencils {
				if stencil.Filename == stencilFile {
					// there is a stencil with the same name. abort
					printFatal("Another stencil with the same name is found. You can use the update command to update it")
					return
				}
			}
			foundFormation = formation
		}
	}

	if err := addStencil(stack, &foundFormation, stencilFile, contextID, template, sequence, message, tags); err != nil {
		printFatal(err.Error())
	}

	fmt.Println("Stencil was added to formation")
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

func printStencil(stencil cloud66.Stencil) {
	var buffer bytes.Buffer

	buffer.WriteString(stencil.Body)
	fmt.Print(buffer.String())
}

func addStencil(stack *cloud66.Stack, formation *cloud66.Formation, stencilFile string, contextID string, templateFilename string, sequence int, message string, tags []string) error {
	body, err := ioutil.ReadFile(stencilFile)
	if err != nil {
		return err
	}

	remoteFilename := filepath.Base(stencilFile)
	stencil := &cloud66.Stencil{
		Filename:         remoteFilename,
		TemplateFilename: templateFilename,
		ContextID:        contextID,
		Tags:             tags,
		Body:             string(body),
		Sequence:         sequence,
	}

	_, err = client.AddStencils(stack.Uid, formation.Uid, []*cloud66.Stencil{stencil}, message)
	if err != nil {
		return err
	}

	return nil
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

type stencilBySequence []cloud66.Stencil

func (a stencilBySequence) Len() int           { return len(a) }
func (a stencilBySequence) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a stencilBySequence) Less(i, j int) bool { return a[i].Sequence < a[j].Sequence }

/* End Stencils */

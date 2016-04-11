package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/cloud66/cli"
	"github.com/cloud66/cloud66"
)

var cmdUsers = &Command{
	Name:       "users",
	Build:      buildUsers,
	NeedsStack: false,
	NeedsOrg:   true,
	Short:      "user and account management actions",
}

func buildUsers() cli.Command {
	base := buildBasicCommand()
	base.Subcommands = []cli.Command{
		cli.Command{
			Name:   "list",
			Usage:  "shows the list of all users in an account",
			Action: runUsers,
			Description: `

Examples:
$ cx users list
Id			Email
1329		jim@gmail.com
2492		jack@gmail.com
`,
		},
		cli.Command{
			Name:   "show",
			Usage:  "returns detiails about a user",
			Action: runShowUser,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "json",
					Usage: "dump as JSON file (with the given name)",
				},
			},
			Description: `
Examples:
$ cx users show jim@gmail.com --json=/tmp/jim_profile.json
`,
		},
		cli.Command{
			Name:   "apply-profile",
			Usage:  "uploads and applies an access profile to a user or a group of users",
			Action: runApplyProfile,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "override",
					Usage: "overrides any existing access profile for the user",
				},
				cli.StringFlag{
					Name:  "json",
					Usage: "profile to be applied to the user in JSON",
				},
			},
			Description: `
Examples:
$ cx users apply-profile jack@gmail.com --json=/tmp/jim_profile.json
`,
		},
	}

	return base
}

func runUsers(c *cli.Context) {
	mustOrg(c)
	users, err := client.ListUsers()
	if err != nil {
		printFatal(err.Error())
	}

	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()

	printUsersList(w, users)
}

func runShowUser(c *cli.Context) {

	var found *cloud66.User
	users, err := client.ListUsers()
	if err != nil {
		printFatal(err.Error())
	}

	if len(c.Args()) != 1 {
		printFatal("invalid usage")
	}

	email := strings.Trim(c.Args().First(), " ")
	for _, user := range users {
		if user.Email == email {
			found = &user
			break
		}
	}

	if found == nil {
		printFatal("Unable to find user %s", email)
		return
	}

	jsonFile := c.String("json")
	user, err := client.GetUser(found.Id)
	if err != nil {
		printFatal(err.Error())
	}

	if jsonFile != "" {
		fmt.Printf("Exporting the profile for %s to %s\n", email, jsonFile)
		b, err := json.MarshalIndent(user.AccessProfile, "", "\t")
		if err != nil {
			printFatal(err.Error())
		}

		manifest, err := os.Create(jsonFile)
		defer manifest.Close()

		manifest.Write(b)
	} else {
		printUser(user)
	}
}

func runApplyProfile(c *cli.Context) {
	mustOrg(c)
	var found *cloud66.User
	users, err := client.ListUsers()
	if err != nil {
		printFatal(err.Error())
	}

	if len(c.Args()) < 1 {
		printFatal("invalid usage")
	}

	email := strings.Trim(c.Args().First(), " ")
	for _, user := range users {
		if user.Email == email {
			found = &user
			break
		}
	}

	if found == nil {
		printFatal("Unable to find user %s", email)
		return
	}

	j := c.String("json")

	var apt cloud66.AccessProfileType
	b, err := ioutil.ReadFile(j)
	if err != nil {
		printFatal("Failed to load %s due to %s", j, err.Error())
	}
	err = json.Unmarshal(b, &apt)
	if err != nil {
		printFatal("Invalid JSON profile %s", err.Error())
	}

	user := cloud66.User{}
	user.AccessProfile = apt
	user.AccessProfile.Override = c.Bool("override")
	_, err = client.UpdateUser(found.Id, user)
	if err != nil {
		printFatal("Failed to upload and apply the profile to %s due to %s", email, err.Error())
	}

	fmt.Println("Profile applied successfully")
}

func printUsersList(w io.Writer, users []cloud66.User) {
	listRec(w,
		"Id",
		"Email",
	)

	for _, user := range users {
		listRec(w,
			user.Id,
			user.Email,
		)
	}
}

func printUser(u *cloud66.User) {
	fmt.Printf("Email: %s\n", u.Email)
	fmt.Printf("Locked: %t\n", u.Locked)
	fmt.Println("Access Profile")
	printAccessProfile(u.AccessProfile)
	fmt.Printf("Uses Two Factor Authentication: %t\n", u.UsesTfa)
	fmt.Printf("Timezone: %s\n", u.Timezone)
	fmt.Printf("Has Valid Phone: %t\n", u.HasValidPhone)
	fmt.Printf("Developer Program: %t\n", u.DeveloperProgram)
	fmt.Printf("Github Login: %t\n", u.GithubLogin)
	fmt.Printf("Last Login: %s\n", u.LastLogin.Local())
	fmt.Printf("Created At: %s\n", u.CreatedAt.Local())
	fmt.Printf("Updated At: %s\n", u.UpdatedAt.Local())
}

func printAccessProfile(a cloud66.AccessProfileType) {
	printAccountProfile(a.AccountProfile)
	for _, r := range a.StackProfiles {
		printStackProfile(r)
	}
	fmt.Println()
}

func printStackProfile(a cloud66.StackProfileType) {
	stack, err := client.FindStackByUid(a.StackUid)
	if err != nil {
		fmt.Errorf("Failed to find stack with UID %s", a.StackUid)
	}

	fmt.Printf("\tStack: %s (%s) - [%s]\n", stack.Name, stack.Environment, a.StackUid)
	fmt.Printf("\t\tRoles: %s\n", a.Role)
}

func printAccountProfile(a cloud66.AccountProfileType) {
	fmt.Printf("\tAccount Profile\n")
	fmt.Printf("\tCanCreateStack: %t\n", a.CanCreateStack)
	fmt.Printf("\tCanAdminUsers: %t\n", a.CanAdminUsers)
	fmt.Printf("\tCanAdminPayments: %t\n", a.CanAdminPayments)
	fmt.Printf("\tCanAddCloudKey: %t\n", a.CanAddCloudKey)
	fmt.Printf("\tCanDelCloudKey: %t\n", a.CanDelCloudKey)
	fmt.Printf("\tCanViewAccountNotifications: %t\n", a.CanViewAccountNotifications)
	fmt.Printf("\tCanEditAccountNotifications: %t\n", a.CanEditAccountNotifications)
	fmt.Printf("\tCanViewAudit: %t\n", a.CanViewAudit)
	fmt.Printf("\tCanViewDockerImageKey: %t\n", a.CanViewDockerImageKey)
	fmt.Printf("\tCanDelSshKey: %t\n", a.CanDelSshKey)
	fmt.Printf("\tCanEditPersonalToken: %t\n", a.CanEditPersonalToken)
	fmt.Printf("\tCanDelAuthorizedApp: %t\n", a.CanDelAuthorizedApp)
	fmt.Printf("\tCanViewCustomEnv: %t\n", a.CanViewCustomEnv)
	fmt.Printf("\tCanEditCustomEnv: %t\n", a.CanEditCustomEnv)
	fmt.Printf("\tCanAddDevelopersApp: %t\n", a.CanAddDevelopersApp)
	fmt.Printf("\tCanDelDevelopersAdd: %t\n", a.CanDelDevelopersAdd)
	fmt.Printf("\tCanEditGitKey: %t\n", a.CanEditGitKey)
	fmt.Printf("\tCanEditGateway: %t\n", a.CanEditGateway)
	fmt.Printf("\tDefault Roles: %v\n", a.DefaultRoles)
	fmt.Println()
}

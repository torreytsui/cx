package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/cloud66/cli"
)

var cmdConfig = &Command{
	Name:       "config",
	Build:      buildConfig,
	NeedsStack: false,
	NeedsOrg:   false,
	Short:      "configuration commands",
}

func buildConfig() cli.Command {
	base := buildBasicCommand()
	base.Subcommands = []cli.Command{
		cli.Command{
			Name:   "list",
			Usage:  "shows the list of all configurations",
			Action: runConfigList,
		},
		cli.Command{
			Name:   "show",
			Usage:  "shows a single configuration values",
			Action: runShowConfig,
		},
		cli.Command{
			Name:   "use",
			Usage:  "switches the configuration to use the given profile",
			Action: runUseConfig,
		},
		cli.Command{
			Name:   "delete",
			Usage:  "delete a given configuration profile",
			Action: runDeleteConfig,
		},
		cli.Command{
			Name:      "create",
			Usage:     "create a new profile",
			ShortName: "create profile_name",
			Action:    runCreateConfig,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "org",
					Usage: "Name of an organization you are a member of. Can be left empty. You can overwrite this using --org argument on other commands",
				},
				cli.StringFlag{
					Name:  "api-url",
					Usage: "URL for Cloud 66 API. Used for OnPrem and Dedicated installations of Cloud 66 Enterprise",
				},
				cli.StringFlag{
					Name:  "base-url",
					Usage: "URL for Cloud 66 Application. Used for OnPrem and Dedicated installations of Cloud 66 Enterprise",
				},
				cli.StringFlag{
					Name:  "client-id",
					Usage: "OAuth Client ID. Used for OnPrem and Dedicated installations of Cloud 66 Enterprise",
				},
				cli.StringFlag{
					Name:  "client-secret",
					Usage: "OAuth Client Secret. Used for OnPrem and Dedicated installations of Cloud 66 Enterprise",
				},
				cli.StringFlag{
					Name:  "faye-endpoint",
					Usage: "URL for realtime push service. Used for OnPrem and Dedicated installations of Cloud 66 Enterprise",
				},
			},
			Description: `
Example:
cx config create foo --org acme
`,
		},
	}

	return base
}

func runConfigList(c *cli.Context) {
	profiles := readProfiles()

	for _, profile := range profiles.Profiles {
		if profile.Name == profiles.LastProfile {
			fmt.Print("* ")
		}

		fmt.Println(profile.Name)
	}
}

func runShowConfig(c *cli.Context) {
	profiles := readProfiles()
	toShow := c.Args().First()
	if toShow == "" {
		printFatal("no profile passed")
	}

	for _, profile := range profiles.Profiles {
		if profile.Name == toShow {
			fmt.Printf("Name: %s\n", profile.Name)
			fmt.Printf("Organization: %s\n", profile.Organization)
			fmt.Println()

			fmt.Printf("ApiURL: %s\n", profile.ApiURL)
			fmt.Printf("BaseURL: %s\n", profile.BaseURL)
			fmt.Printf("FayeEndpoint: %s\n", profile.FayeEndpoint)
			return
		}
	}
}

func runUseConfig(c *cli.Context) {
	toUse := c.Args().First()
	if toUse == "" {
		printFatal("no profile passed")
	}

	profiles := readProfiles()
	if profiles.LastProfile == toUse {
		fmt.Printf("current profile is %s already\n", toUse)
		return
	}

	for _, profile := range profiles.Profiles {
		if profile.Name == toUse {
			profiles.LastProfile = toUse
			if err := profiles.WriteProfiles(); err != nil {
				printFatal("error saving profiles", err)
			}

			fmt.Println("profile switched")
			return
		}
	}

	printFatal("no profile found with the given name")
}

func runDeleteConfig(c *cli.Context) {
	toDelete := c.Args().First()
	if toDelete == "" {
		printFatal("no profile passed")
	}

	profiles := readProfiles()
	findProfile(profiles, toDelete)

	delete(profiles.Profiles, toDelete)

	if len(profiles.Profiles) == 0 {
		if err := os.Remove(profilePath); err != nil {
			printFatal("error deleting profile", err)
		}
	} else {
		if err := profiles.WriteProfiles(); err != nil {
			printFatal("error deleting profile", err)
		}
	}

	fmt.Println("Profile deleted")
}

func runCreateConfig(c *cli.Context) {
	defProfile := defaultProfile()

	name := c.Args().First()

	if name == "" {
		printFatal("no name given")
	}

	if name == "default" {
		printFatal("you cannot create a profile called default")
	}

	org := c.String("org")
	apiURL := c.String("api-url")
	baseURL := c.String("base-url")
	fayeEndpoint := c.String("faye-endpoint")
	clientID := c.String("client-id")
	clientSecret := c.String("client-secret")

	if apiURL == "" {
		apiURL = defProfile.ApiURL
	}
	if baseURL == "" {
		baseURL = defProfile.BaseURL
	}
	if fayeEndpoint == "" {
		fayeEndpoint = defProfile.FayeEndpoint
	}
	if clientID == "" {
		clientID = defProfile.ClientID
	}
	if clientSecret == "" {
		clientSecret = defProfile.ClientSecret
	}

	profile := &Profile{
		ApiURL:       apiURL,
		BaseURL:      baseURL,
		FayeEndpoint: fayeEndpoint,
		Organization: org,
		Name:         name,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenFile:    fmt.Sprintf("cx_%s.json", strings.ToLower(name)),
	}

	profiles := readProfiles()

	for _, p := range profiles.Profiles {
		if p.Name == name {
			printFatal("profile name already exists. Use update or delete instead")
		}
	}

	profiles.Profiles[name] = profile
	if err := profiles.WriteProfiles(); err != nil {
		printFatal("error during saving profiles", err)
	}

	fmt.Println("Profile created")
}

func readProfiles() *Profiles {
	profiles, err := ReadProfiles(profilePath)
	if err != nil {
		printFatal("reading profiles:", err)
	}

	return profiles
}

func findProfile(profiles *Profiles, name string) *Profile {
	for idx, profile := range profiles.Profiles {
		if profile.Name == name {
			return profiles.Profiles[idx]
		}
	}

	printFatal("cannot find the named profile")

	return nil
}

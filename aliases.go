package main

import "fmt"

func populateAliases(commands []*Command) []*Command {

	aliasMap := make(map[string][]string)
	aliasMap["backups"] = []string{"backup"}
	aliasMap["containers"] = []string{"container"}
	aliasMap["databases"] = []string{"database"}
	aliasMap["jobs"] = []string{"job"}
	aliasMap["servers"] = []string{"server"}
	aliasMap["services"] = []string{"service"}
	aliasMap["settings"] = []string{"setting"}
	aliasMap["stacks"] = []string{"stack"}
	aliasMap["processes"] = []string{"process"}

	for key, aliases := range aliasMap {
		for _, command := range commands {
			if command.Name == key {
				for _, alias := range aliases {
					var newCmd = &Command{
						NeedsStack: command.NeedsStack,
						NeedsOrg:   command.NeedsOrg,
						Name:       alias,
						Build:      command.Build,
						Short:      fmt.Sprintf("[%s alias] %s", key, command.Short),
					}
					commands = append(commands, newCmd)
				}
			}
		}
	}
	return commands
}

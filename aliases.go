package main

import "fmt"

func populateAliases(commands []*Command) []*Command {

	aliasMap := make(map[string][]string)
	aliasMap["backups"] = []string{"backup"}
	aliasMap["containers"] = []string{"container"}
	aliasMap["databases"] = []string{"database"}
	aliasMap["servers"] = []string{"server"}
	aliasMap["services"] = []string{"service"}
	aliasMap["settings"] = []string{"setting"}
	aliasMap["stacks"] = []string{"stack"}

	for key, aliases := range aliasMap {
		for _, command := range commands {
			if command.Name == key {
				for _, alias := range aliases {
					var newCmd = &Command{
						NeedsStack: command.NeedsStack,
						Name:       alias,
						Build:      command.Build,
						Short:      fmt.Sprintf("%s [alias]", command.Short),
					}
					commands = append(commands, newCmd)
				}
			}
		}
	}
	return commands
}

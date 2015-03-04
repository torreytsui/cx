package main

import (
	"fmt"

	"github.com/cloud66/cli"
)

func runNewBackup(c *cli.Context) {
	stack := mustStack(c)

	var flagDbTypes *string

	if c.IsSet("dbtypes") {
		flagDbTypes = new(string)
		*flagDbTypes = c.String("dbtypes")
	}

	var flagFrequency *string

	if c.IsSet("frequency") {
		flagFrequency = new(string)
		*flagFrequency = c.String("frequency")
	}

	var flagKeep *int

	if c.IsSet("keep") {
		flagKeep = new(int)
		*flagKeep = c.Int("keep")
	}

	var flagGzip *bool

	if c.IsSet("gzip") {
		flagGzip = new(bool)
		*flagGzip = c.Bool("gzip")
	}

	var flagExcludetables *string

	if c.IsSet("exclude-tables") {
		flagExcludetables = new(string)
		*flagExcludetables = c.String("exclude-tables")
	}

	var flagRunonreplica *bool

	if c.IsSet("run-on-replica") {
		flagRunonreplica = new(bool)
		*flagRunonreplica = c.Bool("run-on-replica")
	}

	err := client.NewBackup(stack.Uid, flagDbTypes, flagFrequency, flagKeep, flagGzip, flagExcludetables, flagRunonreplica)

	if err != nil {
		printFatal("Error during backup creation: " + err.Error())
		return
	}

	fmt.Println("queued for creation")
}

package main

import (
	"github.com/cloud66/cli"
)

var cmdDatabases = &Command{
	Name:       "databases",
	Build:      buildDatabases,
	Short:      "commands to work with databases",
	NeedsStack: true,
	NeedsOrg:   false,
}

func buildDatabases() cli.Command {
	base := buildBasicCommand()
	base.Subcommands = []cli.Command{
		cli.Command{
			Name:   "promote-slave",
			Action: runSlavePromote,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "dbtype",
					Usage: "type of the database",
				},
			},
			Usage: "promotes the specified slave database server to a standalone master",
			Description: `Promotes the specified slave database server to a standalone master.

The slave will be reconfigured as the new standalone DB. The existing master and other existing slaves will be orphaned, and will need to be removed, after which you can scale them up again.

WARNING: This action could result in application downtime, it is advisable to choose a non-busy time to perform this action, and to place your stack in maintenance mode.
In the case of any servers not being accessible during this time, those servers will remain unchanged. It is therefore important to stop/shutdown those servers in this case.
(or to manually stop the DB service on those servers) as having multiple masters in a cluster could cause problems throughout the cluster.

Examples:
$ cx databases promote-slave -s 'my stack name' redis_slave_name
$ cx databases promote-slave -s 'my stack name' --dbtype=postgresql pg_slave1
`,
		},
		cli.Command{
			Name:   "resync-slave",
			Action: runSlaveResync,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "dbtype",
					Usage: "type of the database",
				},
			},
			Usage: "re-syncs the specified slave database server with its master database server",
			Description: `Re-syncs the specified slave database server with its master database server.

From time-to-time your slave db server might go out of sync with its master. This action attempts to re-sync your specified slave server.
This can happen depending on many factors (such as DB size, frequency of change, networking between servers etc)

Examples:
$ cx databases resync-slave -s 'my stack name' postgresql_slave_name
$ cx databases resync-slave -s 'my stack name' --db-type=postgresql pg_slave1
`,
		},
	}

	return base
}

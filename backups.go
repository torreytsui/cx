package main

import (
	"io"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/cloud66/cloud66"

	"github.com/cloud66/cli"
)

var cmdBackups = &Command{
	NeedsStack: true,
	Build:      buildBackups,
	Name:       "backups",
	Short:      "commands to work with backups",
}

func buildBackups() cli.Command {
	base := buildBasicCommand()
	base.Subcommands = []cli.Command{
		cli.Command{
			Name:   "list",
			Usage:  "lists all the managed backups of a stack",
			Action: runBackups,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name: "latest,l",
					Usage: `This will list all the managed backups of a stack grouped by their database type and/or backup schedule
   The list will include backup id, db type, db name, backup status, last activity, restore and verification statuses.
   The -l option will return the latest successful backups.

Examples:
   $ cx backups list
   23212  mysql  mystack_production  Ok  Mar 27 14:00  Not Restored  Not Verified
   23211  redis  mystack_production  Ok  Mar 27 14:00  Not Restored  Not Verified
   34067  mysql  mystack_production  Ok  Mar 27 13:00  Not Restored  Not Verified
   34066  redis  mystack_production  Ok  Mar 27 13:00  Not Restored  Not Verified
   12802  mysql  mystack_production  Ok  Mar 27 12:00  Not Restored  Not Verified
   12801  redis  mystack_production  Ok  Mar 27 12:00  Not Restored  Not Verified

   $ cx backups list --dbtype mysql
   23212  mysql  mystack_production  Ok  Mar 27 14:00  Not Restored  Not Verified
   34067  mysql  mystack_production  Ok  Mar 27 13:00  Not Restored  Not Verified
   12802  mysql  mystack_production  Ok  Mar 27 12:00  Not Restored  Not Verified

   $ cx backups list -latest
   23212  mysql  mystack_production  Ok  Mar 27 14:00  Not Restored  Not Verified
   23211  redis  mystack_production  Ok  Mar 27 14:00  Not Restored  Not Verified

   $ cx backups list -l --dbtype redis
   23211  redis  mystack_production  Ok  Mar 27 14:00  Not Restored  Not Verified
`,
				},
				cli.StringFlag{
					Name:  "dbtype",
					Usage: "Database type",
				},
			},
		},
		cli.Command{
			Name:   "download",
			Action: runDownloadBackup,
			Usage:  "backups download [-d <download directory>] <backup Id>",
			Description: `This downloads a backup from the available backups of a stack. This is limited to a single database type.
The command might download multiple files in parallel and concatenate and untar them if needed. The resulting file
can be used to manually restore the database.

-d allows you to set the directory used to download the backup. You need to have write permissions over that directory
if no directory is specified, ~/cx_backups is used. If the directory does not exist, it will be created.

The caller needs to have admin rights over the stack.

Examples:
$ cx backups download -s mystack 123
`,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "directory,d",
				},
			},
		},
		cli.Command{
			Name:   "new",
			Action: runNewBackup,
			Usage:  "create a new backup task for your stack.",
			Description: `Create a new backup task for your stack.

Example:
$ cx backups new -s mystack	--dbtypes=postgresql --frequency="0 */1 * * *" --gzip=true exclude-tables=my_log_table --run-on-replica=false`,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "dbtypes",
					Usage: "Comma separated list of Database types which need backup tasks i.e mysql,postgresql, ... . Default value is \"all\" ",
				},
				cli.StringFlag{
					Name:  "frequency",
					Usage: "Frequency of backup task in cron schedule format. Put cron string in double quotes i.e \"0 */2 * * *\" .  Default value is \"0 */1 * * *\" ",
				},
				cli.IntFlag{
					Name:  "keep",
					Usage: "Number of previous backups to keep. Default value is 100.",
				},
				cli.BoolFlag{
					Name:  "gzip",
					Usage: "Compress your backups with gzip. Default value is true.",
				},
				cli.StringFlag{
					Name:  "exclude-tables",
					Usage: "Tables that must be excluded from the backup.",
				},
				cli.BoolFlag{
					Name:  "run-on-replica",
					Usage: "Run backup task on replica server if available. Default value is true.",
				},
			},
		},
	}

	return base
}

func runBackups(c *cli.Context) {
	var dbType = c.String("dbtype")

	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()

	stack := mustStack(c)

	backups, err := client.ManagedBackups(stack.Uid)
	must(err)

	var dbTypeGroup = map[string][]cloud66.ManagedBackup{}
	if c.Bool("latest") {
		for _, i := range backups {
			if dbTypeGroup[i.DbType] == nil {
				// it's a new one
				dbTypeGroup[i.DbType] = []cloud66.ManagedBackup{i}
			} else {
				dbTypeGroup[i.DbType] = append(dbTypeGroup[i.DbType], i)
			}
		}

		// now sort each group
		topResults := []cloud66.ManagedBackup{}
		for _, v := range dbTypeGroup {
			sort.Sort(backupsByDate(v))
			topResults = append(topResults, v[0])
		}
		printBackupList(w, topResults, dbType)
	} else {
		printBackupList(w, backups, dbType)
	}
}

func printBackupList(w io.Writer, backups []cloud66.ManagedBackup, dbType string) {
	sort.Sort(backupsByDate(backups))
	for _, a := range backups {
		if dbType == "" || strings.ToLower(a.DbType) == strings.ToLower(dbType) {
			listBackup(w, a)
		}
	}
}

func listBackup(w io.Writer, a cloud66.ManagedBackup) {
	listRec(w,
		a.Id,
		a.DbType,
		a.DatabaseName,
		cloud66.BackupStatus[a.BackupStatus],
		prettyTime{a.BackupDate},
		cloud66.RestoreStatus[a.RestoreStatus],
		cloud66.VerifyStatus[a.VerifyStatus],
	)
}

type backupsByDate []cloud66.ManagedBackup

func (a backupsByDate) Len() int           { return len(a) }
func (a backupsByDate) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a backupsByDate) Less(i, j int) bool { return a[i].BackupDate.Unix() > a[j].BackupDate.Unix() }

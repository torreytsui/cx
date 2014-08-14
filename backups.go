package main

import (
	"io"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/cloud66/cloud66"
)

var cmdBackups = &Command{
	Run:        runBackups,
	Usage:      "backups [-l] [<db type>]",
	NeedsStack: true,
	Category:   "stack",
	Short:      "lists all the managed backups of a stack",
	Long: `This will list all the managed backups of a stack grouped by their database type and/or backup schedule
  The list will include backup id, db type, db name, backup status, last activity, restore and verification statuses.

  -l  will retrun the latest successful backups

  Examples:

  $ cx backups
  23212  mysql  mystack_production  Ok  Mar 27 14:00  Not Restored  Not Verified
  23211  redis  mystack_production  Ok  Mar 27 14:00  Not Restored  Not Verified
  34067  mysql  mystack_production  Ok  Mar 27 13:00  Not Restored  Not Verified
  34066  redis  mystack_production  Ok  Mar 27 13:00  Not Restored  Not Verified
  12802  mysql  mystack_production  Ok  Mar 27 12:00  Not Restored  Not Verified
  12801  redis  mystack_production  Ok  Mar 27 12:00  Not Restored  Not Verified

  $ cx backups mysql
  23212  mysql  mystack_production  Ok  Mar 27 14:00  Not Restored  Not Verified
  34067  mysql  mystack_production  Ok  Mar 27 13:00  Not Restored  Not Verified
  12802  mysql  mystack_production  Ok  Mar 27 12:00  Not Restored  Not Verified

  $ cx backups -l
  23212  mysql  mystack_production  Ok  Mar 27 14:00  Not Restored  Not Verified
  23211  redis  mystack_production  Ok  Mar 27 14:00  Not Restored  Not Verified

  $ cx backups -l redis
  23211  redis  mystack_production  Ok  Mar 27 14:00  Not Restored  Not Verified
`,
}

var (
	flagLatest bool
)

func init() {
	cmdBackups.Flag.BoolVar(&flagLatest, "l", false, "latest successful backup")
}

func runBackups(cmd *Command, args []string) {
	if len(args) > 1 {
		cmd.printUsage()
		os.Exit(2)
	}

	var dbType = ""
	if len(args) == 1 {
		dbType = args[0]
	}

	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()

	stack := mustStack()

	backups, err := client.ManagedBackups(stack.Uid)
	must(err)

	var dbTypeGroup = map[string][]cloud66.ManagedBackup{}
	if flagLatest {
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

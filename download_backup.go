package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

var cmdDownloadBackup = &Command{
	Run:        runDownloadBackup,
	Usage:      "download-backup [-d <download directory>] <backup Id>",
	NeedsStack: false,
	Category:   "stack",
	Short:      "downloads a database backup.",
	Long: `This downloads a backup from the available backups of a stack. This is limited to a single database type.
  The command might download multiple files in parallel and concatenate and untar them if needed. The resulting file
  can be used to manually restore the database.

    -d allows you to set the directory used to download the backup. You need to have write permissions over that directory.
       if no directory is specified, ~/cx_backups is used. If the directory does not exist, it will be created.

  The caller needs to have admin rights over the stack.
  `,
}

var flagDownloadDir string

func init() {
	cmdDownloadBackup.Flag.StringVar(&flagDownloadDir, "d", "", "download directory")
}

func runDownloadBackup(cmd *Command, args []string) {
	if len(args) == 0 {
		cmd.printUsage()
		os.Exit(2)
	}

	backupId, err := strconv.Atoi(args[0])
	if err != nil {
		cmd.printUsage()
		os.Exit(2)
	}

	segment, err := client.GetBackupSegment(backupId, "")
	must(err)

	mainDir := filepath.Join(homePath(), "cx_backups")
	if flagDownloadDir != "" {
		mainDir = flagDownloadDir
	}

	fmt.Println(mainDir)

	// create a download tmp folder
	dir := filepath.Join(mainDir, "tmp", args[0])
	err = os.MkdirAll(dir, 0777)
	must(err)

	var files = []string{}
	for {
		fmt.Printf("Downloading %s to %s...\n", segment.Filename, dir)
		// this should be moved to go routines
		toFile := filepath.Join(dir, segment.Filename)
		err = downloadFile(segment.Url, toFile)
		must(err)

		files = append(files, toFile)

		if segment.NextExtension == "" {
			break
		}

		segment, err = client.GetBackupSegment(backupId, segment.NextExtension)
		must(err)
	}

	toFile := filepath.Join(homePath(), "cx_backups", "backup_"+args[0]+".tar")
	fmt.Printf("Concatenating files to %s\n", toFile)
	err = appendFiles(files, toFile)
	if err != nil {
		printFatal("Error during concatenation: " + err.Error())
		return
	}

	// remove the temp
	if !debugMode {
		os.RemoveAll(dir)
		fmt.Printf("Deleting %s\n", dir)
	}
	fmt.Println("Done")
}

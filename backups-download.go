package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/codegangsta/cli"
)

func runDownloadBackup(c *cli.Context) {
	if len(c.Args()) == 0 {
		//cmd.printUsage()
		os.Exit(2)
	}

	stack := mustStack(c)
	backupId, err := strconv.Atoi(c.Args()[0])
	if err != nil {
		//cmd.printUsage()
		os.Exit(2)
	}

	segmentIndeces, err := client.GetBackupSegmentIndeces(stack.Uid, backupId)
	must(err)
	if len(segmentIndeces) < 1 {
		printFatal("Cannot find file segments associated with this backup")
	}

	flagDownloadDir := c.String("directory")

	mainDir := filepath.Join(homePath(), "cx_backups")
	if flagDownloadDir != "" {
		mainDir = flagDownloadDir
	}

	// create a download tmp folder
	dir := filepath.Join(mainDir, "tmp", c.Args()[0])
	err = os.MkdirAll(dir, 0777)
	must(err)

	var files = []string{}
	for _, segmentIndex := range segmentIndeces {

		segment, err := client.GetBackupSegment(stack.Uid, backupId, segmentIndex.Extension)
		must(err)

		fmt.Printf("Downloading %s to %s\n", segmentIndex.Filename, dir)
		// this should be moved to go routines
		toFile := filepath.Join(dir, segmentIndex.Filename)
		err = downloadFile(segment.Url, toFile)
		must(err)

		files = append(files, toFile)
	}

	toFile := filepath.Join(mainDir, "backup_"+c.Args()[0]+".tar")
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

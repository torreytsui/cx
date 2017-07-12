package main

import (
	"os"
	"time"

	"github.com/cloud66/cloud66"

	"github.com/cloud66/cli"
)

func runJobRun(c *cli.Context) {

	stack := mustStack(c)

	// get the job
	if len(c.Args()) != 1 {
		cli.ShowSubcommandHelp(c)
		os.Exit(2)
	}
	jobName := c.Args()[0]

	jobs, err := client.GetJobs(stack.Uid, nil)
	if err != nil {
		printFatal(err.Error())
	}

	var jobNames []string
	for _, job := range jobs {
		jobNames = append(jobNames, job.GetBasicJob().Name)
	}

	idx, err := fuzzyFind(jobNames, jobName, false)
	if err != nil {
		printFatal(err.Error())
	}
	jobUid := string(jobs[idx].GetBasicJob().Uid)

	jobArgs := ""

	if len(c.StringSlice("arg")) > 0 {
		for i, arg := range c.StringSlice("arg") {
			if i > 0 {
				jobArgs = jobArgs + " "
			}
			jobArgs = jobArgs + "\"" + arg + "\""
		}
	}

	asyncId, err := startJobRun(stack.Uid, jobUid, &jobArgs)
	if err != nil {
		printFatal(err.Error())
	}
	genericRes, err := endJobRun(*asyncId, stack.Uid)
	if err != nil {
		printFatal(err.Error())
	}
	printGenericResponse(*genericRes)
	return
}

func startJobRun(stackUid string, jobId string, jobArgs *string) (*int, error) {
	asyncRes, err := client.RunJobNow(stackUid, jobId, jobArgs)
	if err != nil {
		return nil, err
	}
	return &asyncRes.Id, err
}

func endJobRun(asyncId int, stackUid string) (*cloud66.GenericResponse, error) {
	return client.WaitStackAsyncAction(asyncId, stackUid, 5*time.Second, 20*time.Minute, true)
}

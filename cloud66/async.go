package cloud66

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

const (
	DefaultCheckFrequency = 3 * time.Second   // every 10 seconds
	DefaultTimeout        = 600 * time.Second // 10 minutes
)

type AsyncResult struct {
	Id           int        `json:"id"`
	User         string     `json:"user"`
	ResourceType string     `json:"resource_type"`
	ResourceId   int        `json:"resource_id"`
	Action       string     `json:"action"`
	StartedVia   string     `json:"started_via"`
	StartedAt    time.Time  `json:"started_at"`
	FinishedAt   *time.Time `json:"finished_at"`
	Outcome      *string    `json:"outcome"`
}

func (c *Client) WaitForAsyncActionComplete(uid string, async_result *AsyncResult, err error, checkFrequency time.Duration, timeout time.Duration, showOutput bool) error {
	if err == nil {
		if showOutput {
			fmt.Printf("Executing \"%s\"\n..", async_result.Action)
		}
		var timeoutTime = time.Now().Add(timeout)
		var timedOut = false
		for async_result.FinishedAt == nil && timedOut != true {
			if showOutput {
				fmt.Printf(".")
			}
			// sleep for checkFrequency secs between lookup requests
			time.Sleep(checkFrequency)
			async_result, err = c.getStackAsyncAction(uid, async_result.ResourceType, async_result.ResourceId, async_result.Id)
			timedOut = time.Now().After(timeoutTime)
		}
		if timedOut {
			if showOutput {
				fmt.Println("")
			}
			err = errors.New("timed-out after " + strconv.FormatInt(int64(timeout)/int64(time.Second), 10) + " second(s)")
		} else {
			if showOutput {
				fmt.Println("complete!")
			}
			if async_result.Outcome != nil {
				fmt.Println(async_result.Outcome)
			}
		}
	}
	return err
}

func (c *Client) getStackAsyncAction(uid string, resourceType string, resourceId int, asyncActionId int) (*AsyncResult, error) {
	params := struct {
		ResourceType string `json:"resource_type"`
		ResourceId   int    `json:"resource_id"`
	}{
		ResourceType: resourceType,
		ResourceId:   resourceId,
	}
	req, err := c.NewRequest("GET", "/stacks/"+uid+"/actions/"+strconv.Itoa(asyncActionId)+".json", params)
	if err != nil {
		return nil, err
	}
	var asyncRes *AsyncResult
	return asyncRes, c.DoReq(req, &asyncRes)
}

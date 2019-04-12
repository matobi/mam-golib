package camunda

import (
	"time"

	"github.com/matobi/mam-golib/pkg/errid"
)

// Quit Called to stop loop in function Run().
func (c *Client) Quit() {
	c.isQuit = true
}

// Run Loop to find jobs in Camunda. Loop continues until you call Quit().
func (c *Client) Run(taskChan chan<- Task) {
	pollSleep := time.Second * 30
	timeNextPoll := time.Now().Add(pollSleep)

	for !c.isQuit {
		time.Sleep(1 * time.Second)
		if time.Now().Before(timeNextPoll) {
			continue
		}
		timeNextPoll = time.Now().Add(pollSleep)

		if len(taskChan) > 0 {
			continue
		}

		task, err := c.NextTask()
		if err != nil {
			c.lw.Err("failed call camunda", err)
			continue
		}
		if task == nil {
			continue // no available task
		}

		taskChan <- *task
		timeNextPoll = time.Now().Add(5 * time.Second) // poll again if we found a task
	}
	close(taskChan)
}

// ProcessTasks Receives Tasks from channel and calls matching TaskHandlerFunc.
func (c *Client) ProcessTasks(handlers map[string]TaskHandlerFunc, taskChan <-chan Task) {
	for task := range taskChan {
		c.lw.Info("processTasks begin", "key", task.BusinessKey)
		handler, found := handlers[task.TopicName]
		if !found {
			c.lw.Err("no matching handler", "topic", task.TopicName)
			continue
		}
		taskErr := handler(task)
		externalID := task.Vars.Get("archiveId") // todo: use variable

		if taskErr != nil {
			isTempErr := errid.IsTemporary(taskErr)
			c.lw.Err("bpmTaskFailed", taskErr, "temporary", isTempErr, "externalId", externalID, "procdefKey", task.Key, "activityId", task.ActivityID, "topic", task.TopicName, "businessKey", task.BusinessKey)
			if !isTempErr {
				if err := c.FailTask(task, taskErr); err != nil {
					c.lw.Err("could not terminate failed camunda task", err, "externalId", externalID, "procdefKey", task.Key, "activityId", task.ActivityID, "topic", task.TopicName, "businessKey", task.BusinessKey)
				}
			}
			continue
		}
		c.lw.Info("processTasks done", "key", task.BusinessKey)
		c.lw.Info("bpmTaskDone", "externalId", externalID, "procdefKey", task.Key, "activityId", task.ActivityID, "topic", task.TopicName, "businessKey", task.BusinessKey)
		if err := c.CompleteTask(task); err != nil {
			c.lw.Err("could not complete camunda task", err, "externalId", externalID, "procdefKey", task.Key, "activityId", task.ActivityID, "topic", task.TopicName, "businessKey", task.BusinessKey)
		}
	}
	c.lw.Info("processTasks end")
}

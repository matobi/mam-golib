package camunda

import (
	"fmt"
	"net/http"
	"time"

	"github.com/matobi/mam-golib/pkg/logger"
)

// Client Used to call camunda.
type Client struct {
	host       string
	WorkerID   string
	query      topicQuery
	isQuit     bool
	httpClient *http.Client
	lw         logger.Logger
}

// Task A camunda task that should be handled by a TaskHandlerFunc
type Task struct {
	ActivityID         string
	ActivityInstanceID string
	ID                 string
	Key                string
	TopicName          string
	BusinessKey        string
	Vars               Variables
}

// TaskHandlerFunc Function that proceses a camunda task.
type TaskHandlerFunc func(tast Task) error

// NewClient Creates a new camunda client.
func NewClient(host string, workerID string, topics []string, variables []string, delay time.Duration, httpClient *http.Client, lw logger.Logger) *Client {
	return &Client{
		host:       host,
		WorkerID:   workerID,
		query:      getTopicQuery(workerID, topics, variables, delay),
		httpClient: httpClient,
		lw:         lw,
	}
}

// NextTask Returns next available task or nil if none available.
func (c *Client) NextTask() (*Task, error) {
	url := c.urlPoll()
	tasks := []jsonTask{}
	if err := c.callCamunda(http.MethodPost, url, c.query, &tasks); err != nil {
		return nil, err // no connection to camunda.
	}
	taskCount := len(tasks)
	if taskCount == 0 {
		return nil, nil // no tasks
	}
	if taskCount > 1 {
		return nil, fmt.Errorf("expected single task, got multiple. len=%d", taskCount)
	}

	task := &Task{
		ActivityID:         tasks[0].ActivityID,
		ActivityInstanceID: tasks[0].ActivityInstanceID,
		ID:                 tasks[0].ID,
		Key:                tasks[0].Key,
		TopicName:          tasks[0].TopicName,
		BusinessKey:        tasks[0].BusinessKey,
		Vars:               getVariables(tasks[0]),
	}
	return task, nil
}

// CompleteTask Called when a task has been successfully handled.
// Will complete task in camunda.
func (c *Client) CompleteTask(task Task) error {
	url := c.urlComplete(task.ID)
	complete := completeMsg{
		WorkerID:  c.WorkerID,
		Variables: updateTaskVariables(task.Vars),
	}
	return c.callCamunda(http.MethodPost, url, complete, nil)
}

// FailTask Called when a task has failed.
// If error is marked as temporary (see package errid) then the task will be retried later.
// If error is not temporary, then the task will create an incident in camunda.
func (c *Client) FailTask(task Task, err error) error {
	failure := c.newFailure(err)
	url := c.urlFail(task.ID)
	return c.callCamunda(http.MethodPost, url, failure, nil)
}

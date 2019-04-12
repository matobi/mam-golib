package camunda

import (
	"fmt"

	"github.com/matobi/mam-golib/pkg/ws"
)

// Rest API to camunda.

type jsonTask struct {
	ActivityID         string `json:"activityId"`
	ActivityInstanceID string `json:"activityInstanceId"`
	ID                 string `json:"id"`
	Key                string `json:"processDefinitionKey"`
	TopicName          string `json:"topicName"`
	BusinessKey        string `json:"businessKey"`

	Vars map[string]jsonVariable `json:"variables"`
}

type jsonVariable struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

func (c *Client) callCamunda(method, url string, in interface{}, out interface{}) error {
	return ws.NewCaller(method, url).JSON().Call(c.httpClient, in, out)
}

func (c *Client) urlPoll() string {
	return fmt.Sprintf("%s/engine-rest/external-task/fetchAndLock", c.host)
}

func (c *Client) urlComplete(taskID string) string {
	return fmt.Sprintf("%s/engine-rest/external-task/%s/complete", c.host, taskID)
}

func (c *Client) urlFail(taskID string) string {
	return fmt.Sprintf("%s/engine-rest/external-task/%s/failure", c.host, taskID)
}

func (c *Client) newFailure(err error) failMsg {
	retries := 0
	// todo: not sure how to handle retires in camunda.
	//if errid.IsTemporary(err) {
	//	retries = 5
	//}
	return failMsg{
		WorkerID:     c.WorkerID,
		ErrorMsg:     err.Error(),
		Retries:      retries,
		RetryTimeout: 600 * 1000,
	}
}

type failMsg struct {
	WorkerID     string `json:"workerId"`
	ErrorMsg     string `json:"errorMessage"`
	Retries      int    `json:"retries"`
	RetryTimeout int64  `json:"retryTimeout"`
}

type completeMsg struct {
	WorkerID  string                  `json:"workerId"`
	Variables map[string]jsonVariable `json:"variables"`
}

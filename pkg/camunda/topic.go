package camunda

import (
	"time"
)

// Filter to find tasks in Camunda

type topicQuery struct {
	WorkerID  string        `json:"workerId"`
	MaxTasks  int           `json:"maxTasks"`
	UsePrio   bool          `json:"usePriority"`
	Variables []string      `json:"variables"`
	Topics    []topicFilter `json:"topics"`
}

type topicFilter struct {
	Name   string `json:"topicName"`
	LockMs int64  `json:"lockDuration"`
}

func getTopicQuery(workerID string, topics []string, variables []string, delay time.Duration) topicQuery {
	filter := []topicFilter{}
	for _, topic := range topics {
		f := topicFilter{
			Name:   topic,
			LockMs: delay.Nanoseconds() / 1000000,
		}
		filter = append(filter, f)
	}
	q := topicQuery{
		WorkerID:  workerID,
		MaxTasks:  1,
		UsePrio:   true,
		Variables: variables,
		Topics:    filter,
	}
	return q
}

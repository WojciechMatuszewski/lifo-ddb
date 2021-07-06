package task

import (
	"context"
	"encoding/json"
	"fmt"
	"lifo-ddb/src/cfg"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchevents"
	ebTypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchevents/types"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/pkg/errors"
)

func RunTask(ctx context.Context, task Task) error {
	buf, err := json.Marshal(task)
	if err != nil {
		return errors.Wrap(err, "failed to marshal the task item to be run")
	}

	config := cfg.New()
	client := cloudwatchevents.NewFromConfig(config)

	out, err := client.PutEvents(ctx, &cloudwatchevents.PutEventsInput{
		Entries: []ebTypes.PutEventsRequestEntry{
			{
				Detail:     aws.String(string(buf)),
				Source:     aws.String("TaskProcessor"),
				DetailType: aws.String("API Call via TaskProcessor"),
			},
		},
	})
	if err != nil {
		return errors.Wrap(err, "failed to send the event")
	}

	if out.FailedEntryCount > 0 {
		failedEvt := out.Entries[0]
		return fmt.Errorf("failed to send the event, code: %v, message: %v", *failedEvt.ErrorCode, *failedEvt.ErrorMessage)
	}

	return nil
}

type Status string

const (
	Task_STATUS_PENDING Status = "PENDING"
	Task_STATUS_TAKEN   Status = "TAKEN"
	Task_STATUS_SUCCESS Status = "SUCCESS"
)

type Task struct {
	ID        string    `json:"id,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
	UpdatedAt time.Time `json:"updatedAt,omitempty"`
	Status    Status    `json:"status,omitempty"`
}

func (t Task) ToItem() TaskItem {
	return TaskItem{
		ID:        t.ID,
		CreatedAt: t.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: t.UpdatedAt.UTC().Format(time.RFC3339),
		Status:    string(t.Status),
		TTL:       t.CreatedAt.Add(time.Minute * 10).Unix(),
	}
}

type TaskItem struct {
	ID        string `dynamodbav:"id"`
	CreatedAt string `dynamodbav:"createdAt"`
	UpdatedAt string `dynamodbav:"updatedAt"`
	Status    string `dynamodbav:"status"`
	TTL       int64  `dynamodbav:"ttl"`
}

func (ti TaskItem) ToTask() Task {
	createdAt, err := time.Parse(time.RFC3339, ti.CreatedAt)
	if err != nil {
		panic(err)
	}

	updatedAt, err := time.Parse(time.RFC3339, ti.UpdatedAt)
	if err != nil {
		panic(err)
	}

	return Task{
		ID:        ti.ID,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Status:    Status(ti.Status),
	}
}

func FromAvs(avs map[string]types.AttributeValue) (TaskItem, error) {
	item := &TaskItem{}
	err := attributevalue.UnmarshalMap(avs, item)
	if err != nil {
		return TaskItem{}, errors.Wrap(err, "failed to unmarshal the avs")
	}

	return *item, nil
}

func (ti TaskItem) ToAvs() (map[string]types.AttributeValue, error) {

	avs, err := attributevalue.MarshalMap(ti)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal the the task item")
	}

	return avs, nil
}

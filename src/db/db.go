package db

import (
	"context"
	"lifo-ddb/src/task"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type DB struct {
	client    *dynamodb.Client
	logger    *log.Logger
	tableName string
}

func NewDB(logger *log.Logger, client *dynamodb.Client, tableName string) *DB {
	return &DB{logger: logger, client: client, tableName: tableName}
}

func (db *DB) CreateTask(ctx context.Context) error {
	db.logger.Println("Creating task")
	t := task.Task{
		ID:        uuid.New().String(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Status:    task.Task_STATUS_PENDING,
	}
	db.logger.Printf("Task created %#v \n", t)

	avs, err := t.ToItem().ToAvs()
	if err != nil {
		return err
	}

	db.logger.Printf("Avs that will be saved %#v", avs)
	_, err = db.client.PutItem(ctx, &dynamodb.PutItemInput{
		Item:      avs,
		TableName: aws.String(db.tableName),
	})
	if err != nil {
		return errors.Wrap(err, "failed to write the task")
	}
	db.logger.Printf("Task saved %#v \n", t)

	return nil
}

func (db *DB) GetTasks(ctx context.Context) ([]task.Task, error) {
	db.logger.Println("Getting tasks")

	out, err := db.client.Query(ctx, &dynamodb.QueryInput{
		TableName: aws.String(db.tableName),
		ExpressionAttributeNames: map[string]string{
			"#status": "status",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":status": &types.AttributeValueMemberS{
				Value: string(task.Task_STATUS_PENDING),
			},
		},
		Limit:                  aws.Int32(10),
		IndexName:              aws.String("byStatus"),
		KeyConditionExpression: aws.String("#status = :status"),
		ScanIndexForward:       aws.Bool(false),
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch tasks")
	}

	db.logger.Printf("Retrived items #%v", out.Items)

	tasks := make([]task.Task, len(out.Items))
	for i, avs := range out.Items {
		t, err := task.FromAvs(avs)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert the av")
		}

		tasks[i] = t.ToTask()
	}

	db.logger.Printf("Retrieved tasks %#v", tasks)

	return tasks, nil
}

func (db *DB) TransitionTask(ctx context.Context, tsk task.Task, from task.Status, to task.Status) (task.Task, error) {
	db.logger.Printf("Updating task %#v from: %v to %v", tsk, from, to)

	avs, err := tsk.ToItem().ToAvs()
	if err != nil {
		return task.Task{}, nil
	}

	db.logger.Printf("Trying to update the task with a key %#v", avs)

	out, err := db.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		Key: map[string]types.AttributeValue{
			"id": avs["id"],
		},
		TableName:           aws.String(db.tableName),
		ConditionExpression: aws.String("#status = :fromStatus"),
		ExpressionAttributeNames: map[string]string{
			"#status":    "status",
			"#updatedAt": "updatedAt",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":toStatus": &types.AttributeValueMemberS{
				Value: string(to),
			},
			":updatedAt": &types.AttributeValueMemberS{
				Value: time.Now().Format(time.RFC3339),
			},
			":fromStatus": &types.AttributeValueMemberS{
				Value: string(from),
			},
		},
		ReturnValues:     types.ReturnValueAllNew,
		UpdateExpression: aws.String("SET #status = :toStatus, #updatedAt = :updatedAt"),
	})
	if err != nil {
		return task.Task{}, errors.Wrap(err, "failed to update the item")
	}

	tskItem, err := task.FromAvs(out.Attributes)
	if err != nil {
		return task.Task{}, err
	}

	db.logger.Printf("Task after transitioning %#v", tskItem.ToTask())

	return tskItem.ToTask(), nil
}

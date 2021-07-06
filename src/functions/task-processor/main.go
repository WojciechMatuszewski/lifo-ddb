package main

import (
	"context"
	"lifo-ddb/src/cfg"
	"lifo-ddb/src/db"
	"lifo-ddb/src/task"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

func main() {
	logger := log.New(os.Stdout, "[Task Processor] ", log.LstdFlags)

	config := cfg.New()
	db := db.NewDB(logger, dynamodb.NewFromConfig(config), os.Getenv("TABLE_NAME"))

	lambda.Start(newHandler(logger, db))
}

type Handler func(ctx context.Context) error

func newHandler(logger *log.Logger, db *db.DB) Handler {
	return func(ctx context.Context) error {
		tasks, err := db.GetTasks(ctx)
		if err != nil {
			logger.Println(err)

			return err
		}

		logger.Printf("Tasks to be processed %#v \n", tasks)

		for _, tsk := range tasks {
			upTsk, err := db.TransitionTask(ctx, tsk, tsk.Status, task.Task_STATUS_TAKEN)
			if err != nil {
				logger.Println(err)
				return err
			}

			err = task.RunTask(ctx, upTsk)
			if err != nil {
				logger.Println(err)
				return err
			}

			_, err = db.TransitionTask(ctx, upTsk, upTsk.Status, task.Task_STATUS_SUCCESS)
			if err != nil {
				logger.Println(err)
				return err
			}
		}

		return nil
	}
}

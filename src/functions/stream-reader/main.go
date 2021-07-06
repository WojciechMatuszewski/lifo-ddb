package main

import (
	"context"
	"lifo-ddb/src/cfg"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

func main() {
	lambda.Start(newHandler(log.New(os.Stdout, "[Stream Reader] ", log.LstdFlags)))
}

type Handler func(ctx context.Context, event events.DynamoDBEvent) error

func newHandler(logger *log.Logger) Handler {
	return func(ctx context.Context, event events.DynamoDBEvent) error {
		logger.Printf("records: %#v \n", event.Records)

		hasInsertRecords := false
		for _, record := range event.Records {
			if record.EventName == string(events.DynamoDBOperationTypeInsert) {
				hasInsertRecords = true
			}
		}

		if !hasInsertRecords {
			logger.Println("Nothing to publish, skipping")
			return nil
		}

		config := cfg.New()
		client := sns.NewFromConfig(config)

		topicArn := os.Getenv("PROCESS_TASKS_TOPIC")
		logger.Printf("Sending to topic %#v \n", topicArn)

		_, err := client.Publish(ctx, &sns.PublishInput{
			Message:  aws.String("There are tasks for processign"),
			TopicArn: aws.String(topicArn),
		})
		if err != nil {
			logger.Print(err)

			return err
		}

		return nil
	}
}

package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	lambda.Start(newHandler(log.New(os.Stdout, "[Failure Destination] ", log.LstdFlags)))
}

type Handler func(ctx context.Context, event interface{}) error

func newHandler(logger *log.Logger) Handler {
	return func(ctx context.Context, event interface{}) error {
		logger.Printf("event %#v", event)

		return nil
	}
}

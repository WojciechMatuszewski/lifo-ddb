package main

import (
	"context"
	"lifo-ddb/src/cfg"
	"lifo-ddb/src/db"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

func main() {
	config := cfg.New()
	logger := log.New(os.Stdout, "[Create Task] ", log.LstdFlags)

	db := db.NewDB(logger, dynamodb.NewFromConfig(config), os.Getenv("TABLE_NAME"))

	lambda.Start(newHandler(db))
}

type Handler func(ctx context.Context) error

func newHandler(db *db.DB) Handler {
	return func(ctx context.Context) error {
		return db.CreateTask(ctx)
	}
}

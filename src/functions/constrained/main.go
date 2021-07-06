package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

const url = "https://webhook.site/8339ac80-7755-49c2-8cd4-4d8c5d2af899"

func main() {
	lambda.Start(newHandler(log.New(os.Stdout, "[Constrainted Function] ", log.LstdFlags)))
}

type Handler func(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayProxyResponse, error)

func newHandler(logger *log.Logger) Handler {
	return func(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayProxyResponse, error) {
		fmt.Println("Waiting")

		time.Sleep(time.Second * 1)

		fmt.Println("Request", event.Body)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer([]byte(event.Body)))
		if err != nil {
			return events.APIGatewayProxyResponse{Body: http.StatusText(http.StatusInternalServerError), StatusCode: http.StatusInternalServerError}, err
		}

		req.Header.Set("Content-Type", "text/plain")
		out, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println(err)
			return events.APIGatewayProxyResponse{Body: http.StatusText(http.StatusInternalServerError), StatusCode: http.StatusInternalServerError}, err
		}
		defer out.Body.Close()

		fmt.Println(out.StatusCode)

		if out.StatusCode != http.StatusOK {
			return events.APIGatewayProxyResponse{Body: http.StatusText(out.StatusCode), StatusCode: out.StatusCode}, nil
		}

		return events.APIGatewayProxyResponse{Body: http.StatusText(http.StatusOK), StatusCode: http.StatusOK}, nil
	}
}

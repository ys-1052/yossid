package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	echoadapter "github.com/awslabs/aws-lambda-go-api-proxy/echo"
	"github.com/ys-1052/yossid/backend/internal/app"
)

var adapter *echoadapter.EchoLambdaV2

func init() {
	log.Println("Initializing yossid OIDC Provider (AWS Lambda Mode)...")
	ctx := context.Background()

	application, err := app.NewApp(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	adapter = echoadapter.NewV2(application.Echo)
}

func Handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	// Proxy request to Echo
	return adapter.ProxyWithContext(ctx, req)
}

func main() {
	lambda.Start(Handler)
}

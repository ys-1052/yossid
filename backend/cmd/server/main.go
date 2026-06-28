package main

import (
	"context"
	"log"

	"github.com/ys-1052/yossid/backend/internal/app"
)

func main() {
	ctx := context.Background()
	log.Println("Starting yossid OIDC Provider (Local HTTP Mode)...")

	application, err := app.NewApp(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	log.Printf("Listening on port %s", application.Config.Port)
	if err := application.Start(); err != nil {
		log.Fatalf("Application shut down with error: %v", err)
	}
}

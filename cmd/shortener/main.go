package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/GlebRadaev/shlink/internal/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	application := app.NewApplication(ctx)
	if err := application.Init(); err != nil {
		log.Fatalf("Application initialization error: %v", err)
	}
	if err := application.Start(); err != nil {
		log.Fatalf("Application runtime error: %v", err)
	}
}

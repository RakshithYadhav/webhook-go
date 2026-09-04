package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/RakshithYadhav/webhook-go/internal/event"
	"github.com/RakshithYadhav/webhook-go/internal/intake"
	"github.com/RakshithYadhav/webhook-go/internal/registry"
	"github.com/RakshithYadhav/webhook-go/internal/worker"
	"github.com/google/uuid"
)

func main() {
	testData := map[string][]string{
		"insert": {
			"https://customer-a.example.com/hooks/schedule-events",
			"https://customer-a.example.com/hooks/backup",
		},
	}

	testEvent := event.Event{
		ResourceID: uuid.Max,
		EventType:  "insert",
		Payload:    json.RawMessage{},
	}

	wg := sync.WaitGroup{}

	register := registry.Registry{}

	register.Seed(testData)

	intakeService := intake.New(&register, &wg)

	err := intakeService.Submit(testEvent)
	if err != nil {
		fmt.Printf("Shutdown")
	}

	client := worker.New()
	pool := worker.NewPool(intakeService.Queue(), 3, *client, &wg)
	pool.Start()

	// user story 4 shutdown mechanism.
	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM)

	// Ctx done returns a open channel which is blocked if no one listens.
	// when internally the shutdown signal is got ctx.Done channel 
	// will be unblocked there by triggering the shutdown mechanism.
	<-ctx.Done()

	shCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	intakeService.Shutdown()
	poolErr := pool.Shutdown(shCtx)
	if poolErr != nil {
		fmt.Printf("Shutdown")
	}
	defer stop()

}

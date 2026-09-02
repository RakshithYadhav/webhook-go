package main

import (
	"encoding/json"
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

	register := registry.Registry{}

	register.Seed(testData)

	intakeService := intake.New(&register)

	intakeService.Submit(testEvent)

	client := worker.New()
	pool := worker.NewPool(intakeService.Queue(), 3, *client)
	pool.Start()

	// Temporary stand-in until issue #4 (graceful shutdown) gives main() a real
	// reason to keep running. select{} looked safer but actually deadlocks here:
	// once this one-shot batch is delivered, every goroutine (main included) is
	// permanently parked with nothing left to ever wake any of them, and Go's
	// runtime correctly kills the process for it. Sleeping is fine here because
	// there's no ongoing concurrent work to synchronize with — just letting one
	// known, bounded, already-in-flight batch finish before exiting.
	time.Sleep(3 * time.Second)
}

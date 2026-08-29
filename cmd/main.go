package main

import (
	"encoding/json"

	"github.com/RakshithYadhav/webhook-go/internal/event"
	"github.com/RakshithYadhav/webhook-go/internal/intake"
	"github.com/RakshithYadhav/webhook-go/internal/registry"
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
}

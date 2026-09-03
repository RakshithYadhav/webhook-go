package intake

import (
	"encoding/json"
	"github.com/RakshithYadhav/webhook-go/internal/event"
	"github.com/RakshithYadhav/webhook-go/internal/registry"
	"github.com/google/uuid"
	"sync"
	"testing"
)

func TestSubmitFanOut(t *testing.T) {
	// arrange
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

	// call
	register := registry.Registry{}
	wg := sync.WaitGroup{}
	register.Seed(testData)
	intakeService := New(&register, &wg)

	err := intakeService.Submit(testEvent)
	// assert
	if err != nil {
		t.Fatalf("Submit returned unexpected error : %v", err)
	}

	wantLen := len(testData["insert"])
	if got := len(intakeService.queue); got != wantLen {
		t.Errorf("queue length = %d, want %d", got, wantLen)
	}
}

func TestSubmitFanOutWithZeroEndpoints(t *testing.T) {
	// arrange
	testData := map[string][]string{
		"insert": {
			"https://customer-a.example.com/hooks/schedule-events",
			"https://customer-a.example.com/hooks/backup",
		},
	}

	testEvent := event.Event{
		ResourceID: uuid.Max,
		EventType:  "update",
		Payload:    json.RawMessage{},
	}

	// call
	register := registry.Registry{}
	wg := sync.WaitGroup{}
	register.Seed(testData)
	intakeService := New(&register, &wg)

	err := intakeService.Submit(testEvent)

	// assert
	if err != ErrNoEndPoints {
		t.Fatalf("Expecting No Endpoints Error : %v", err)
	}

	wantLen := 0

	if got := len(intakeService.queue); got != wantLen {
		t.Errorf("queue length = %d, want %d", got, wantLen)
	}
}

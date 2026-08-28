package event

import (
	"github.com/google/uuid"
	"encoding/json"
)

type Event struct {
	ResourceId uuid.UUID
	EventType string
	Payload json.RawMessage
}
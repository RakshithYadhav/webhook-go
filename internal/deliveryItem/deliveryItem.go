package deliveryitem

import (
	"github.com/RakshithYadhav/webhook-go/internal/event"
)

type DeliveryItem struct {
	Event    event.Event
	Endpoint string
}

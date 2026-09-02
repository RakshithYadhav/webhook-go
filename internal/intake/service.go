package intake

import (
	"errors"
	"github.com/RakshithYadhav/webhook-go/internal/deliveryItem"
	"github.com/RakshithYadhav/webhook-go/internal/event"
	"github.com/RakshithYadhav/webhook-go/internal/registry"
)

const queueCapacity = 100

var ErrNoEndPoints = errors.New("No Endpoint for this event")

type IntakeService struct {
	queue    chan (deliveryitem.DeliveryItem)
	register *registry.Registry
}

func New(register *registry.Registry) *IntakeService {
	return &IntakeService{
		queue:    make(chan deliveryitem.DeliveryItem, queueCapacity),
		register: register,
	}
}

func (i *IntakeService) Submit(event event.Event) error {
	endpoints := i.register.Lookup(event.EventType)

	if len(endpoints) == 0 {
		return ErrNoEndPoints
	}

	for _, endpoint := range endpoints {
		item := deliveryitem.DeliveryItem{
			Event:    event,
			Endpoint: endpoint,
		}

		i.queue <- item
	}

	return nil
}

// Returns <-chan, not chan: queue itself stays private so only Submit can send.
// The receive-only return type is what actually restricts callers to reading —
// Submit still holds full send access via the private field underneath.
func (i *IntakeService) Queue() <-chan (deliveryitem.DeliveryItem) {
	return i.queue
}

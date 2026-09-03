package intake

import (
	"errors"
	"github.com/RakshithYadhav/webhook-go/internal/deliveryItem"
	"github.com/RakshithYadhav/webhook-go/internal/event"
	"github.com/RakshithYadhav/webhook-go/internal/registry"
	"sync"
	"sync/atomic"
)

const queueCapacity = 100

var ErrNoEndPoints = errors.New("No Endpoint for this event")
var ErrShutdownError = errors.New("Shutdown intiated, no more events will be accepted")

type IntakeService struct {
	queue            chan (deliveryitem.DeliveryItem)
	register         *registry.Registry
	shutDownIntiated atomic.Bool
	wg               *sync.WaitGroup
}

func New(register *registry.Registry, wg *sync.WaitGroup) *IntakeService {
	return &IntakeService{
		queue:    make(chan deliveryitem.DeliveryItem, queueCapacity),
		register: register,
		wg:       wg,
	}
}

func (i *IntakeService) Submit(event event.Event) error {
	if i.shutDownIntiated.Load() {
		return ErrShutdownError
	}

	endpoints := i.register.Lookup(event.EventType)

	if len(endpoints) == 0 {
		return ErrNoEndPoints
	}

	for _, endpoint := range endpoints {
		i.wg.Add(1)
		item := deliveryitem.DeliveryItem{
			Event:    event,
			Endpoint: endpoint,
		}

		i.queue <- item
	}

	return nil
}

func (i *IntakeService) Shutdown() {
	i.shutDownIntiated.Store(true)
}

// Returns <-chan, not chan: queue itself stays private so only Submit can send.
// The receive-only return type is what actually restricts callers to reading —
// Submit still holds full send access via the private field underneath.
func (i *IntakeService) Queue() <-chan (deliveryitem.DeliveryItem) {
	return i.queue
}

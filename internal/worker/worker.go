package worker

import (
	"bytes"
	"net/http"
	"time"

	"github.com/RakshithYadhav/webhook-go/internal/deliveryItem"
	"github.com/RakshithYadhav/webhook-go/internal/resourceLock"
)

const timeOut = 2 * time.Second

type WorkerClient struct {
	// A dedicated client, not http.Get/Post's DefaultClient, so we can set our own
	// timeout — DefaultClient has none. Connection pooling itself lives in the
	// Transport (shared globally via http.DefaultTransport when Transport is left
	// nil), not in the Client value, so that isn't the reason for holding this here —
	// it's just to avoid rebuilding the timeout config on every call.
	client http.Client
}

type Pool struct {
	eventQueue   <-chan deliveryitem.DeliveryItem
	poolSize     int
	client       WorkerClient
	resourceLock *resourcelock.ResourceLock
}

func NewPool(eventQueue <-chan deliveryitem.DeliveryItem, poolSize int, workerClient WorkerClient) *Pool {
	return &Pool{
		poolSize:     poolSize,
		eventQueue:   eventQueue,
		client:       workerClient,
		resourceLock: resourcelock.NewResourceLock(),
	}
}

func (p *Pool) Start() {
	for worker := 0; worker < p.poolSize; worker += 1 {
		go func() {
			for item := range p.eventQueue {
				// Recovery is scoped per item via this inner function call: defer
				// only runs when its *enclosing function* returns, so recovering
				// at the loop level would exit the whole goroutine on the first
				// panic, not just skip the one bad item.
				func() {
					p.resourceLock.Lock(item.Event.ResourceID)
					defer p.resourceLock.Unlock(item.Event.ResourceID)
					
					defer func() {
						recover()
					}()
					p.client.SendDeliveryItem(item)
				}()
			}

		}()
	}
}

func New() *WorkerClient {
	httpClient := http.Client{
		Timeout: timeOut,
	}
	return &WorkerClient{
		client: httpClient,
	}
}

func (w *WorkerClient) SendDeliveryItem(item deliveryitem.DeliveryItem) {
	res, err := w.client.Post(item.Endpoint, "application/json", bytes.NewReader(item.Event.Payload))
	if err != nil {
		return
	}

	defer res.Body.Close()
}

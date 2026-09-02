package worker

import (
	"bytes"
	"net/http"
	"time"

	"github.com/RakshithYadhav/webhook-go/internal/deliveryItem"
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
	eventQueue <-chan deliveryitem.DeliveryItem
	poolSize   int
	client     WorkerClient
}

func NewPool(eventQueue <-chan deliveryitem.DeliveryItem, poolSize int, workerClient WorkerClient) *Pool {
	return &Pool{
		poolSize:  poolSize,
		eventQueue: eventQueue,
		client:     workerClient,
	}
}

func (p *Pool) Start() {
	for worker := 0; worker < p.poolSize; worker += 1 {
		go func() {
			for item := range p.eventQueue {
				p.client.SendDeliveryItem(item)
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
	w.client.Post(item.Endpoint, "application/json", bytes.NewReader(item.Event.Payload))
}

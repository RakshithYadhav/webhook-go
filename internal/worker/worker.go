package worker

import (
	"bytes"
	"net/http"
	"sync"
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
	eventQueue chan (deliveryitem.DeliveryItem)
	queueSize  int
	client     WorkerClient
	wg sync.WaitGroup
}

func NewPool(eventQueue chan (deliveryitem.DeliveryItem), queueSize int, workerClient WorkerClient) *Pool {
	return &Pool{
		queueSize:  queueSize,
		eventQueue: eventQueue,
		client:     workerClient,
	}
}

func (p *Pool) Start() {
	for worker := 0; worker < p.queueSize; worker += 1 {
		go func() {
			for item := range p.eventQueue {
				p.wg.Add(1)
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

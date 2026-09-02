package worker

import (
	"bytes"
	"net/http"
	"time"
	"github.com/RakshithYadhav/webhook-go/internal/deliveryItem"
)

const timeOut = 2 * time.Second

type Worker struct {
	// A dedicated client, not http.Get/Post's DefaultClient, so we can set our own
	// timeout — DefaultClient has none. Connection pooling itself lives in the
	// Transport (shared globally via http.DefaultTransport when Transport is left
	// nil), not in the Client value, so that isn't the reason for holding this here —
	// it's just to avoid rebuilding the timeout config on every call.
	client http.Client
}

func New() *Worker {
	httpClient := http.Client{
		Timeout: timeOut,
	}
	return &Worker{
		client: httpClient,
	}
}

func (w *Worker) SendDeliveryItem(item deliveryitem.DeliveryItem) {
	w.client.Post(item.Endpoint, "application/json", bytes.NewReader(item.Event.Payload))
}

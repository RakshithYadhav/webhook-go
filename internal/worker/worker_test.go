package worker

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	deliveryitem "github.com/RakshithYadhav/webhook-go/internal/deliveryItem"
	"github.com/RakshithYadhav/webhook-go/internal/event"
	"github.com/RakshithYadhav/webhook-go/internal/intake"
	"github.com/RakshithYadhav/webhook-go/internal/registry"
	"github.com/google/uuid"
)

func TestPostBehaviourOfWorker(t *testing.T) {
	// setup
	didRequestArrive := false

	handler := func(w http.ResponseWriter, r *http.Request) {
		didRequestArrive = true
	}
	// convert handler to handler func which then would have serveHttp.
	// then NewServer which needs something which satisfies go's http.handler interface.
	// the only reason for the Http.HandlerFunc wrapping is to make your plain function quality as an http.handler.
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	// call
	item := deliveryitem.DeliveryItem{
		Endpoint: server.URL,
		Event:    event.Event{},
	}

	worker := New()
	worker.SendDeliveryItem(item)

	// assert
	if !didRequestArrive {
		t.Fatalf("Request Did not arrive")
	}

}

func TestTimeOut(t *testing.T) {
	// set up

	hander := func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
	}

	server := httptest.NewServer(http.HandlerFunc(hander))
	defer server.Close()

	// call
	item := deliveryitem.DeliveryItem{
		Endpoint: server.URL,
		Event:    event.Event{},
	}

	worker := New()

	start := time.Now()
	worker.SendDeliveryItem(item)
	duration := time.Since(start)

	// check
	if duration > 2500*time.Millisecond {
		t.Fatalf("Too slow its slower than 2 seconds")
	}
}

func TestWorkerPool(t *testing.T) {
	// set up.
	// registry -  need to seed and create few endpoints.
	// for test. 3 events with a single endpoint.
	var wg sync.WaitGroup
	handler := func(w http.ResponseWriter, r *http.Request) {
		defer wg.Done()
	}

	server := newServer(handler)

	endpoints := map[string][]string{
		"insert":    {server.URL},
		"update":    {server.URL},
		"processed": {server.URL},
	}

	reg := registry.Registry{}
	reg.Seed(endpoints)

	// Next Intake
	service := intake.New(&reg)
	// Add test events to the queue.
	for index := 0; index < 5; index++ {

		for k, _ := range endpoints {
			sampleEvent := event.Event{
				ResourceID: uuid.New(),
				EventType:  k,
				Payload:    []byte(fmt.Sprintf("index : %d", index)),
			}

			wg.Add(1)
			service.Submit(sampleEvent)
		}
	}

	queue := service.Queue()
	w := New()

	for worker := 0; worker < 3; worker++ {
		go func() {
			for item := range queue {
				w.SendDeliveryItem(item)
			}
		}()
	}

	wg.Wait()
}

func newServer(handler func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(handler))
}

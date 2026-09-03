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
	service := intake.New(&reg, &wg)
	// Add test events to the queue.
	for index := 0; index < 5; index++ {

		for k, _ := range endpoints {
			sampleEvent := event.Event{
				ResourceID: uuid.New(),
				EventType:  k,
				Payload:    []byte(fmt.Sprintf("index : %d", index)),
			}

			service.Submit(sampleEvent)
		}
	}

	client := New()
	pool := NewPool(service.Queue(), 3, *client, &wg)
	pool.Start()

	wg.Wait()
}

func TestResourceLockOnWorkerPool(t *testing.T) {
	// set up.
	// registry -  need to seed and create few endpoints.
	// for test. 3 events with a single endpoint.

	var wg sync.WaitGroup
	var mu sync.Mutex
	inFlight := false
	overLap := false

	handler := func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		if inFlight {
			overLap = true
		} else {
			inFlight = true
		}
		mu.Unlock()

		time.Sleep(time.Millisecond * 200)

		mu.Lock()
		inFlight = false
		mu.Unlock()
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
	service := intake.New(&reg, &wg)
	// Add test events to the queue.
	resId := uuid.New()
	for index := 0; index < 5; index++ {

		for k, _ := range endpoints {
			sampleEvent := event.Event{
				ResourceID: resId,
				EventType:  k,
				Payload:    []byte(fmt.Sprintf("index : %d", index)),
			}

			service.Submit(sampleEvent)
		}
	}

	client := New()
	pool := NewPool(service.Queue(), 3, *client, &wg)
	pool.Start()
	wg.Wait()

	if overLap {
		t.Fatalf("Over lap detected test failed.")
	}
}

func TestDifferentResourcesDeliverConcurrently(t *testing.T) {
	// Different resource IDs must not be serialized against each other by the
	// keyed lock — only same-resource items should ever block one another.
	var wg sync.WaitGroup
	var mu sync.Mutex
	current := 0
	maxConcurrent := 0

	handler := func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		current++
		if current > maxConcurrent {
			maxConcurrent = current
		}
		mu.Unlock()

		time.Sleep(time.Millisecond * 200)

		mu.Lock()
		current--
		mu.Unlock()
	}

	server := newServer(handler)

	endpoints := map[string][]string{
		"insert": {server.URL},
	}

	reg := registry.Registry{}
	reg.Seed(endpoints)

	service := intake.New(&reg, &wg)

	// Interleaved submission (one event per resource per round), not all of one
	// resource's events followed by all of another's — otherwise the queue's
	// FIFO order would hand workers the same resource's items first and this
	// test could pass by accident of ordering rather than by proving anything.
	resourceIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	for round := 0; round < 5; round++ {
		for _, resID := range resourceIDs {
			sampleEvent := event.Event{
				ResourceID: resID,
				EventType:  "insert",
				Payload:    []byte(fmt.Sprintf("round : %d", round)),
			}

			service.Submit(sampleEvent)
		}
	}

	client := New()
	pool := NewPool(service.Queue(), 3, *client, &wg)
	pool.Start()
	wg.Wait()

	if maxConcurrent <= 1 {
		t.Fatalf("Expected deliveries for different resources to overlap, but max concurrent was %d", maxConcurrent)
	}
}

func newServer(handler func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(handler))
}

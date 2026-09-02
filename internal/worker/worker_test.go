package worker

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	deliveryitem "github.com/RakshithYadhav/webhook-go/internal/deliveryItem"
	"github.com/RakshithYadhav/webhook-go/internal/event"
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

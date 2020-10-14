package mock

import (
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestActionHandler(t *testing.T) {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("GET", "/status", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Action)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body is what we expect.
	expected := `{"alive": true}`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestRateLimit(t *testing.T) {
	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.

	rps := 10
	burst := 1
	server := httptest.NewServer(RateLimit(rps, burst, time.Duration(1) * time.Second, Action))
	defer server.Close()
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	counter := int32(0)
	nRequests := 100
	var wg sync.WaitGroup
	start := time.Now()
	// 100 rqps
	for nRequests >= 0 && time.Since(start) <= time.Duration(1) * time.Second{
		nRequests--
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()

			res, err := http.Get(server.URL)
			if err != nil {
				t.Fatal(err)
			}
			if status := res.StatusCode; status == http.StatusOK {
				atomic.AddInt32(&counter, 1)

			} else {
				log.Printf("handler returned wrong status code: got %v want %v",
					status, http.StatusOK)
			}
		}(&wg)

	}

	wg.Wait()
	if counter != int32(rps + burst) {
		t.Errorf("throttling not working: got %v want %v",
			counter, rps + burst)
	}

}

package mock

import (
	"bytes"
	"crypto/tls"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/http/httptrace"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestActionHandler(t *testing.T) {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("POST", "/status", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	action := Action{
		listOfTargets: []string{""},
		currTarget:    0,
		respSize: 1024,
		reqSize: 2048,
	}
	handler := http.HandlerFunc(action.Execute)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body is what we expect.
	if len(rr.Body.String()) != action.respSize {
		t.Errorf("handler returned a body different from expected, body len: %d != expected len: %d", len(rr.Body.String()), action.reqSize)
	}

}

func TestRateLimit(t *testing.T) {
	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.

	rps := 10
	burst := 1
	action := Action{
		listOfTargets: []string{""},
		currTarget:    0,
	}
	server := httptest.NewServer(RateLimit(rps, burst, time.Duration(1) * time.Second, action.Execute))
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

			res, err := http.Post(server.URL,
				"application/octet-stream",
				// create a message of 4KB size
				bytes.NewBufferString(""))
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

func TestLatency(t *testing.T) {
	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.

	rps := 10
	burst := 1
	action := Action{
		listOfTargets: []string{""},
		currTarget:    0,
		respTime:      time.Duration(10) * time.Second,
	}
	server := httptest.NewServer(RateLimit(rps, burst, time.Duration(1) * time.Second, action.Execute))
	defer server.Close()
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	counter := int32(0)
	counterSlow := int32(0)
	nRequests := 100
	var wg sync.WaitGroup
	start := time.Now()
	// 100 rqps
	for nRequests >= 0 && time.Since(start) <= time.Duration(1) * time.Second{
		nRequests--
		wg.Add(1)
		go func(wg *sync.WaitGroup, a* Action) {
			uid := rand.Int()
			defer wg.Done()
			//res, err := http.Post(server.URL,
			//	"application/octet-stream",
			//	// create a message of 4KB size
			//	bytes.NewBufferString(""))

			req, _ := http.NewRequest("POST", server.URL, bytes.NewBufferString(""))

			var start, connect, dns, tlsHandshake time.Time

			trace := &httptrace.ClientTrace{
				DNSStart: func(dsi httptrace.DNSStartInfo) { dns = time.Now() },
				DNSDone: func(ddi httptrace.DNSDoneInfo) {
					log.Printf("[%d] DNS Done: %v\n", uid, time.Since(dns))
				},

				TLSHandshakeStart: func() { tlsHandshake = time.Now() },
				TLSHandshakeDone: func(cs tls.ConnectionState, err error) {
					log.Printf("[%d] TLS Handshake: %v\n", uid, time.Since(tlsHandshake))
				},

				ConnectStart: func(network, addr string) { connect = time.Now() },
				ConnectDone: func(network, addr string, err error) {
					log.Printf("[%d] Connect time: %v\n", uid, time.Since(connect))
				},

				GotFirstResponseByte: func() {
					log.Printf("[%d] Time from start to first byte: %v\n", uid, time.Since(start))
				},
			}

			req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
			start = time.Now()
			res, err := http.DefaultTransport.RoundTrip(req)
			if err != nil {
				log.Fatal(err)
			}

			totalTime := time.Since(start)
			if totalTime >= time.Duration(10) * time.Second {
				atomic.AddInt32(&counterSlow, 1)
			}

			status := res.StatusCode
			if status == http.StatusOK {
				atomic.AddInt32(&counter, 1)
			}
			log.Printf("[%d] handler returned status code: got %v want %v in %v",
				uid, status, http.StatusOK, totalTime)

		}(&wg, &action)

	}

	wg.Wait()

	if counter != int32(rps + burst) {
		t.Errorf("throttling not working: got %v want %v",
			counter, rps + burst)
	}

	if counter != counterSlow {
		t.Errorf("number of slow requests is different of #requests served: got %v want %v",
			counter, counterSlow)
	}

}

func TestChainingRequests(t *testing.T) {
	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rps := 10
	burst := 1

	action2 := Action{
		listOfTargets: []string{""},
		currTarget:    0,
		reqSize: 2,
		respSize: 31,
	}
	server2 := httptest.NewServer(RateLimit(rps, burst, time.Duration(1) * time.Second, action2.Execute))
	defer server2.Close()

	action := Action{
		listOfTargets: []string{server2.URL},
		currTarget:    0,
		reqSize: 5,
		respSize: 37,
	}
	server := httptest.NewServer(RateLimit(rps, burst, time.Duration(1) * time.Second, func(w http.ResponseWriter, req *http.Request) {
		action.Execute(w, req)
	}))
	defer server.Close()

	res, err := http.Post(server.URL,
		"application/octet-stream",
		bytes.NewBufferString(""))
	if err != nil {
		t.Fatal(err)
	}
	if status := res.StatusCode; status != http.StatusOK {
		log.Printf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	if len(body) != action.respSize + action2.respSize {
		t.Fatalf("not appending messages")
	}

}
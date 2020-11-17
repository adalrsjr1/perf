package mock

import (
	"bytes"
	"context"
	"fmt"
	"golang.org/x/time/rate"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"perf/util"
	"time"
)



// https://rodaine.com/2017/05/x-files-time-rate-golang/
func SetRqpsLoad(port uint, rpq, burst, wait int, action* Action) {
	run(port, rpq, burst, wait, action)
}

func run(port uint, rpq, burst, wait int, action* Action) {
	http.HandleFunc("/status", RateLimit(rpq, burst, time.Duration(wait) * time.Millisecond, action.Execute))
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatalf("error when running the server %v", err)
	}
}

type Action struct {
	listOfTargets []string
	currTarget int
	reqSize int
	respSize int
}

func (a* Action) nextAddr() string {
	if len(a.listOfTargets) == 0 {
		return ""
	}
	nextTarget := a.listOfTargets[a.currTarget]
	a.currTarget += (a.currTarget + 1) % len(a.listOfTargets)
	return nextTarget

}

func (a* Action) Execute(w http.ResponseWriter, req *http.Request) {

	if "POST" != req.Method {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	addr := a.nextAddr()
	var body []byte
	var err error
	status := http.StatusOK
	if addr != "" {
		hostname, err := os.Hostname()
		if err != nil {
			log.Printf("cannot fetch hostname setting empty name")
			hostname = ""
		}
		log.Printf("[%v] calling next: %v", hostname, addr)
		resp, err := http.Post(addr,
			"application/octet-stream",
			// create a message of 4KB size
			bytes.NewBufferString(util.String(a.reqSize)))
		if err != nil {
			log.Printf("error '%v' when seding request to '%v'", err, addr)
			status = http.StatusServiceUnavailable
		}
		defer resp.Body.Close()
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("error '%v' when reading requqest", err)
			status = http.StatusInternalServerError
		}
	}

	newBody := []byte(util.String(a.respSize))
	body = append(body, newBody...)

	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	// In the future we could report back on the status of our DB, or our cache
	// (e.g. Redis) by performing a simple PING, and include them in the response.
	_, err = io.WriteString(w, string(body))
	if err != nil {
		log.Printf("error '%v' when writing output", err)
	}
}

// RateLimit middleware limits the throughput to h using a rate.Limiter
// token bucket configured with the provided rps and burst. The request
// will idle for up to the passed in wait. If the limiter detects the
// deadline will be exceeded, the request is cancelled immediately.
func RateLimit(rps, burst int, wait time.Duration, h http.HandlerFunc) http.HandlerFunc {
	l := rate.NewLimiter(rate.Limit(rps), burst)

	return func(w http.ResponseWriter, r *http.Request) {
		// create a new context from the request with the wait timeout
		ctx, cancel := context.WithTimeout(r.Context(), wait)
		defer cancel() // always cancel the context!

		// Wait errors out if the request cannot be processed within
		// the deadline. This is preemptive, instead of waiting the
		// entire duration.
		if err := l.Wait(ctx); err != nil {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}

		h(w, r)
	}
}


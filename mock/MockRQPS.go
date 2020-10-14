package mock

import (
	"context"
	"fmt"
	"golang.org/x/time/rate"
	"io"
	"log"
	"net/http"
	"time"
)
// https://rodaine.com/2017/05/x-files-time-rate-golang/
func SetRqpsLoad(port uint, rpq, burst, wait int) {
	run(port, rpq, burst, wait)
}

func run(port uint, rpq, burst, wait int) {
	http.HandleFunc("/status", RateLimit(rpq, burst, time.Duration(wait) * time.Millisecond, Action))
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatalf("error when running the server %v", err)
	}
}

func Action(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	// In the future we could report back on the status of our DB, or our cache
	// (e.g. Redis) by performing a simple PING, and include them in the response.
	_, err := io.WriteString(w, `{"alive": true}`)
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


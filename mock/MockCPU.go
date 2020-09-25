package mock

import (
	"log"
	"runtime"
	"sync"
	"time"
)

func SetCpuLoad(load float64, duration uint) {
	numThreadsCore := runtime.NumCPU()
	var wg sync.WaitGroup
	for i := 0; i < numThreadsCore; i++ {
		wg.Add(1)
		log.Printf("spawning thread %d\n", i)
		go FinityCpuUsage(i, load, duration, &wg)
	}
	wg.Wait()
}

// https://caffinc.github.io/2016/03/cpu-load-generator/
// set CPU usage in $load% for $timeElapsed ms
func FinityCpuUsage(thread int, load float64, duration uint, wg *sync.WaitGroup) time.Duration {
	defer wg.Done()
	start := time.Now()
	var elapsed time.Duration
	log.Printf("[Thread %d] load:%f duration:%d\n", thread, load, duration)
	for {
		unladenTime := time.Now().UnixNano() / int64(time.Millisecond)
		if unladenTime%100 == 0 {
			time.Sleep(time.Millisecond * time.Duration((1-load)*100))
		}
		elapsed = time.Now().Sub(start)
		if duration > 0 && uint(elapsed/time.Millisecond) >= duration {
			log.Printf("[Thread %d] Ending cpu load\n", thread)
			break
		}
	}
	return elapsed
}

package mock

import (
	"log"
	"runtime"
	"time"
)

func SetMemUsage(megabytes uint64, duration uint) {

	var overall [][]byte
	var i uint64
	for ; i < megabytes; i++ {
		// Mega
		a := make([]byte, 0, 1048576*32)
		overall = append(overall, a)
		memUsage(megabytes)
		if i % 100 == 0 {
			time.Sleep(time.Millisecond * time.Duration(10))
		}
	}
	memUsage(megabytes)
	time.Sleep(time.Millisecond * time.Duration(duration))
}

func memUsage(expected uint64) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	log.Printf("Expected:%v\tAlloc:%vMiB\tTotalAlloc: %vMiB\tSys:%v\tNumGC:%v\n",expected, BToMb(m.Alloc), BToMb(m.TotalAlloc), BToMb(m.Sys), m.NumGC)
}

func BToMb(b uint64) uint64 {
	return b / (1048576*32)
}

func MbToB(mb uint64) uint64 {
	return mb * (1048576*32)
}

package main

import (
	"flag"
	"fmt"
	"os"
	"perf/mock"
)

func main() {

	cpuCmd := flag.NewFlagSet("cpu", flag.ExitOnError)
	cpuLoadArg := cpuCmd.Float64("load", 0.5, "cpu load value (0.0, 1.0)")
	cpuDurationArg := cpuCmd.Uint("duration", 1000, "duration in milliseconds")

	memCmd := flag.NewFlagSet("mem", flag.ExitOnError)
	memLoadArg := memCmd.Uint64("load", 256, "memory load in megabytes")
	memDurationArg := memCmd.Uint("duration", 1000, "duration in milliseconds")

	if len(os.Args) < 2 {
		fmt.Println("expected 'cpu' or 'mem' subcommands")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "cpu":
		cpuCmd.Parse(os.Args[2:])
		fmt.Println("calling 'cpu'")
		mock.SetCpuLoad(*cpuLoadArg, *cpuDurationArg)
	case "mem":
		memCmd.Parse(os.Args[2:])
		fmt.Println("calling 'mem'")
		mock.SetMemUsage(*memLoadArg, *memDurationArg)
	default:
		fmt.Println("expected 'cpu' or 'mem' subcommands")
		os.Exit(1)
	}

}

package main
# to refer: https://hub.docker.com/r/polinux/stress
import (
	"flag"
	"fmt"
	"os"
	"perf/mock"
	"time"
)

func main() {

	cpuCmd := flag.NewFlagSet("cpu", flag.ExitOnError)
	cpuLoadArg := cpuCmd.Float64("load", 0.5, "cpu load value (0.0, 1.0)")
	cpuDurationArg := cpuCmd.Uint("duration", 1000, "duration in milliseconds")

	memCmd := flag.NewFlagSet("mem", flag.ExitOnError)
	memLoadArg := memCmd.Uint64("load", 256, "memory load in megabytes")
	memDurationArg := memCmd.Uint("duration", 1000, "duration in milliseconds")

	rqpsCmd := flag.NewFlagSet("rqps", flag.ExitOnError)
	rqpsLoadArg := rqpsCmd.Int("load", 100, "max number of requests per second")
	rqpsBurstArg := rqpsCmd.Int("burst", 1, "extra requests in addition to the load")
	rqpsWaitArg := rqpsCmd.Int("wait", 1, "waiting time  in millisecond before cancelling a request")
	serverPort := rqpsCmd.Uint("port", 8888, "server port")
	rqpsTargets := rqpsCmd.String("targets", "", "list of targets separated with ;")
	reqSize := rqpsCmd.Int("request-size", 1024, "request size in bytes, the real value is request size * 4")
	resSize := rqpsCmd.Int("response-size", 1024, "response size in bytes, the real value is request size * 4")
	respTime := rqpsCmd.Uint("response-time", 0, "response time in milliseconds")

	if len(os.Args) < 2 {
		fmt.Println("expected either 'cpu', 'mem', 'rqps' subcommands")
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
	case "rqps":
		rqpsCmd.Parse(os.Args[2:])
		fmt.Println("calling 'rqps")

		action := mock.NewAction(*rqpsTargets, *reqSize, *resSize, time.Duration(*respTime) * time.Millisecond)

		mock.SetRqpsLoad(*serverPort, *rqpsLoadArg, *rqpsBurstArg, *rqpsWaitArg, &action)
	default:
		fmt.Println("expected 'cpu', 'mem', or 'rqps' subcommands")
		os.Exit(1)
	}

}

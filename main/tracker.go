package main

import (
	"log"
	"os"
	"time"

	"github.com/remerge/tracker"
)

func main() {
	metadata := &tracker.EventMetadata{
		Service:     "tracker",
		Environment: "development",
		Cluster:     "local",
		Host:        "localhost",
		Release:     "master",
	}

	logger := log.New(os.Stdout, "[sarama] ", log.LstdFlags)
	lt := tracker.NewLogTracker(logger, metadata)
	lt.FastMessage("test", []byte("message"))

	kt, err := tracker.NewKafkaTracker([]string{"0.0.0.0:9092"}, metadata)
	if err != nil {
		panic(err)
	}

	kt.FastMessage("test", []byte("fast message"))
	kt.SafeMessage("test", []byte("safe message"))

	time.Sleep(1 * time.Second)
	kt.Close()
}

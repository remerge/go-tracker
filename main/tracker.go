package main

import (
	"fmt"
	"log"
	"os"
	"time"

	timestr "github.com/remerge/go-timestr"
	"github.com/remerge/go-tracker"
	rand "github.com/remerge/go-xorshift"
)

type TestEvent struct {
	tracker.EventBase
	MyField string
}

func main() {
	fmt.Println(rand.Int63())

	metadata := &tracker.EventMetadata{
		Service:     "tracker",
		Environment: "development",
		Cluster:     "local",
		Host:        "localhost",
		Release:     "master",
	}

	event := &TestEvent{
		MyField: timestr.ISO8601(),
	}

	logger := log.New(os.Stdout, "[sarama] ", log.LstdFlags)
	lt := tracker.NewLogTracker(logger, metadata)
	lt.FastMessage("test", event)

	kt, err := tracker.NewKafkaTracker([]string{"0.0.0.0:9092"}, metadata)
	if err != nil {
		panic(err)
	}

	kt.FastMessage("test", event)
	kt.SafeMessage("test", event)

	time.Sleep(1 * time.Second)
	kt.Close()
}

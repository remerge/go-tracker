package main

import (
	"time"

	"github.com/bobziuchkovski/cue"
	"github.com/bobziuchkovski/cue/collector"
	"github.com/remerge/go-tracker"
)

type TestEvent struct {
	tracker.EventBase
	MyField string
}

var log = cue.NewLogger("main")

func main() {
	cue.Collect(cue.DEBUG, collector.Terminal{}.New())

	metadata := &tracker.EventMetadata{
		Service:     "tracker",
		Environment: "development",
		Cluster:     "local",
		Host:        "localhost",
		Release:     "master",
	}

	event := &TestEvent{
		MyField: "foo",
	}

	kt, err := tracker.NewKafkaTracker([]string{"0.0.0.0:9092"}, metadata)
	if err != nil {
		log.Panic(err, "failed to create kafka tracker")
	}

	kt.FastMessage("test", event)
	kt.SafeMessage("test", event)

	// give background worker a chance to send safe messages
	time.Sleep(1 * time.Second)

	kt.Close()
}

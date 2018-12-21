package main

import (
	"github.com/remerge/cue"
	"github.com/remerge/cue/collector"
	"github.com/remerge/go-tracker"
)

type testEvent struct {
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

	event := &testEvent{
		MyField: "foo",
	}

	kt, err := tracker.NewKafkaTracker([]string{"0.0.0.0:9092"}, metadata)
	if err != nil {
		log.Panic(err, "failed to create kafka tracker")
	}

	// #nosec
	_ = log.Error(kt.FastMessage("test", event), "failed to send fast message")
	_ = log.Error(kt.FastMessageWithKey("test", event, []byte("key")), "failed to send fast message")

	// #nosec
	_ = log.Error(kt.SafeMessage("test", event), "failed to send safe message")
	_ = log.Error(kt.SafeMessageWithKey("test", event, []byte("key")), "failed to send fast message")

	kt.Close()
}

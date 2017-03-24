package tracker

import (
	"encoding/json"

	"github.com/remerge/timestr"
)

// Tracker implements a simple interface to track (unstructured) messages or
// (structured) events.
type Tracker interface {
	Close()
	FastMessage(topic string, message interface{}) error
	SafeMessage(topic string, message interface{}) error
}

// BaseTracker is a base class for implementing trackers
type BaseTracker struct {
	Metadata *EventMetadata
}

func (t *BaseTracker) Encode(message interface{}) ([]byte, error) {
	switch message.(type) {
	case []byte:
		return message.([]byte), nil
	case Event:
		event := message.(Event)
		event.SetMetadata(t.Metadata)
		event.SetTimestamp(timestr.ISO8601())
		return json.Marshal(event)
	case map[string]interface{}:
		event := message.(map[string]interface{})
		event["ts"] = timestr.ISO8601()
		return json.Marshal(event)
	default:
		return json.Marshal(message)
	}
}

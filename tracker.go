package tracker

import (
	"encoding/json"

	"github.com/remerge/cue"
	"github.com/remerge/go-timestr"
)

var log = cue.NewLogger("tracker")

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

// Encode a message into bytes using `json.Marshal` after updating metadata
// fields.
func (t *BaseTracker) Encode(message interface{}) ([]byte, error) {
	switch m := message.(type) {
	case []byte:
		return m, nil
	case string:
		return []byte(m), nil
	case Event:
		if t.Metadata != nil {
			m.SetMetadata(t.Metadata)
		}
		m.SetTimestamp(timestr.ISO8601())
		return m.MarshalJSON()
	case map[string]interface{}:
		if _, found := m["ts"]; !found {
			m["ts"] = timestr.ISO8601()
		}
		return json.Marshal(m)
	default:
		return json.Marshal(m)
	}
}

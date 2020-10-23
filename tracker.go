package tracker

import (
	"encoding/json"
	"time"

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
	FastMessageWithKey(topic string, message interface{}, key []byte) error
	SafeMessageWithKey(topic string, message interface{}, key []byte) error
	CheckHealth() error
}

type Timestampable interface {
	SetTimestampFrom(time.Time, string)
}

// BaseTracker is a base class for implementing trackers
type BaseTracker struct {
	Metadata *EventMetadata
}

// Encode a message into bytes, updating metadata if `Metadatable`, updating the
// timestamp if `Timestampable`  and as json if implementing `json.Marshaler`. Falling back to `json.Marshal` otherwise
// fields.
func (t *BaseTracker) Encode(message interface{}) ([]byte, error) {
	// tombstone
	if message == nil {
		return nil, nil
	}

	// basic types
	switch m := message.(type) {
	case []byte:
		return m, nil
	case string:
		return []byte(m), nil
	case map[string]interface{}:
		if _, found := m["ts"]; !found {
			m["ts"] = timestr.ISO8601()
		}
		return json.Marshal(m)
	}
	if v, ok := message.(Timestampable); ok {
		v.SetTimestampFrom(timestr.NowUTC(), timestr.ISO8601inUTC())
	}
	if v, ok := message.(MetadatableSimple); ok {
		v.SetMetadata(t.Metadata.Service, t.Metadata.Environment, t.Metadata.Cluster,
			t.Metadata.Host, t.Metadata.Release)
	}
	if v, ok := message.(Metadatable); ok {
		v.SetMetadata(t.Metadata)
	}
	if v, ok := message.(json.Marshaler); ok {
		return v.MarshalJSON()
	}
	// fallback
	return json.Marshal(message)
}

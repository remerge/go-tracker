package tracker

import (
	"encoding/json"
)

// Event is a generic interface for accepting structured messages that have
// metadata and can be serialized.
type Event interface {
	SetMetadata(metadata *EventMetadata)
	SetTimestamp(ts string)
	MarshalJSON() ([]byte, error)
}

// EventMetadata represents fields in the Event that are set automatically by
// the tracker for all processed events.
type EventMetadata struct {
	Service     string `json:"service,omitempty"`
	Environment string `json:"env,omitempty"`
	Cluster     string `json:"cluster,omitempty"`
	Host        string `json:"host,omitempty"`
	Release     string `json:"release,omitempty"`
}

// EventBase can be used to implement the Event interface.
type EventBase struct {
	Ts string `form:"ts" json:"ts,omitempty"`
	EventMetadata
}

var _ Event = (*EventBase)(nil)

// SetTimestamp sets the timestamp of the message/event.
func (eb *EventBase) SetTimestamp(ts string) {
	eb.Ts = ts
}

// MarshalJSON marshals the event using the Go default json encoder
func (eb *EventBase) MarshalJSON() ([]byte, error) {
	return json.Marshal(Event(eb))
}

// SetMetadata sets all empty fields in EventBase to values supplied by
// EventMetadata.
func (eb *EventBase) SetMetadata(metadata *EventMetadata) {
	if eb.Service == "" {
		eb.Service = metadata.Service
	}

	if eb.Environment == "" {
		eb.Environment = metadata.Environment
	}

	if eb.Cluster == "" {
		eb.Cluster = metadata.Cluster
	}

	if eb.Host == "" {
		eb.Host = metadata.Host
	}

	if eb.Release == "" {
		eb.Release = metadata.Release
	}
}

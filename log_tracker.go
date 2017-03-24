package tracker

import "log"

// LogTracker is a tracker with in-memory storage used for testing.
type LogTracker struct {
	BaseTracker
	logger *log.Logger
}

var _ Tracker = (*LogTracker)(nil)

// NewLogTracker creates a new mock tracker for testing.
func NewLogTracker(logger *log.Logger, metadata *EventMetadata) *LogTracker {
	t := &LogTracker{}
	t.Metadata = metadata
	t.logger = logger
	return t
}

// Close the tracker.
func (t *LogTracker) Close() {
}

// FastMessage sends a message without waiting for confirmation.
func (t *LogTracker) FastMessage(topic string, message interface{}) error {
	buf, err := t.Encode(message)
	if err != nil {
		return err
	}
	t.logger.Printf("FastMessage(%s, %s)", topic, string(buf))
	return nil
}

// SafeMessage sends a message and waits for confirmation.
func (t *LogTracker) SafeMessage(topic string, message interface{}) error {
	buf, err := t.Encode(message)
	if err != nil {
		return err
	}
	t.logger.Printf("SafeMessage(%s, %s)", topic, string(buf))
	return nil
}

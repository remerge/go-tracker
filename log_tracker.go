package tracker

import "github.com/remerge/cue"

// LogTracker is a tracker with in-memory storage used for testing.
type LogTracker struct {
	BaseTracker
	logger cue.Logger
}

var _ Tracker = (*LogTracker)(nil)

// NewLogTracker creates a new mock tracker for testing.
func NewLogTracker(name string, metadata *EventMetadata) *LogTracker {
	t := &LogTracker{}
	t.Metadata = metadata
	t.logger = cue.NewLogger(name)
	return t
}

// Close the tracker.
func (t *LogTracker) Close() {
}

// FastMessage sends a message without waiting for confirmation.
func (t *LogTracker) FastMessage(topic string, message interface{}) error {
	return t.logMessage("fast", topic, message)
}

// SafeMessage sends a message and waits for confirmation.
func (t *LogTracker) SafeMessage(topic string, message interface{}) error {
	return t.logMessage("safe", topic, message)
}

func (t *LogTracker) logMessage(
	typ string,
	topic string,
	message interface{},
) error {
	buf, err := t.Encode(message)
	if err != nil {
		return err
	}

	t.logger.WithFields(cue.Fields{
		"topic":   topic,
		"message": string(buf),
	}).Info(typ + " message")

	return nil
}

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
func (t *LogTracker) FastMessage(topic string, value interface{}) error {
	return t.logMessage("fast", topic, value, nil)
}

func (t *LogTracker) FastMessageWithKey(topic string, value interface{}, key []byte) error {
	return t.logMessage("fast", topic, value, key)
}

// SafeMessage sends a message and waits for confirmation.
func (t *LogTracker) SafeMessage(topic string, value interface{}) error {
	return t.logMessage("safe", topic, value, nil)
}

func (t *LogTracker) SafeMessageWithKey(topic string, value interface{}, key []byte) error {
	return t.logMessage("safe", topic, value, key)
}

func (t *LogTracker) logMessage(
	typ string,
	topic string,
	value interface{},
	key []byte,
) error {
	valueBuf, err := t.Encode(value)
	if err != nil {
		return err
	}

	if t.logger.EnabledFor(cue.INFO) {
		t.logger.WithFields(cue.Fields{
			"topic": topic,
			"value": string(valueBuf),
			"key":   string(key),
		}).Infof("%s message", typ)
	}

	return nil
}

func (t *LogTracker) CheckHealth() error {
	return nil
}

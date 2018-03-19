package tracker

// MockTracker is a tracker with in-memory storage used for testing.
type MockTracker struct {
	BaseTracker
	Messages map[string][][]byte
}

var _ Tracker = (*MockTracker)(nil)

// NewMockTracker creates a new mock tracker for testing.
func NewMockTracker(metadata *EventMetadata) *MockTracker {
	t := &MockTracker{}
	t.Metadata = metadata
	t.Messages = make(map[string][][]byte)
	return t
}

// Get a message from the mock broker.
func (t *MockTracker) Get(topic string, idx int) []byte {
	if len(t.Messages) == 0 {
		return nil
	}

	msgs := t.Messages[topic]
	if msgs == nil || len(msgs) < idx {
		return nil
	}

	return msgs[idx]
}

// Close the tracker.
func (t *MockTracker) Close() {
}

// FastMessage sends a message without waiting for confirmation.
func (t *MockTracker) FastMessage(topic string, message interface{}) error {
	buf, err := t.Encode(message)
	if err != nil {
		return err
	}
	t.Messages[topic] = append(t.Messages[topic], buf)
	return nil
}

// SafeMessage sends a message and waits for confirmation.
func (t *MockTracker) SafeMessage(topic string, message interface{}) error {
	return t.FastMessage(topic, message)
}

func (t *MockTracker) CheckHealth() error {
	return nil
}

package tracker

// MockTracker is a tracker with in-memory storage used for testing.
type MockTracker struct {
	BaseTracker
	Messages map[string][][]byte
	ids      map[string][][]byte
}

var _ Tracker = (*MockTracker)(nil)

// NewMockTracker creates a new mock tracker for testing.
func NewMockTracker(metadata *EventMetadata) *MockTracker {
	t := &MockTracker{}
	t.Metadata = metadata
	t.Messages = make(map[string][][]byte)
	t.ids = make(map[string][][]byte)
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

// GetKey Gets a key from the mock broker.
func (t *MockTracker) GetKey(topic string, idx int) []byte {
	if len(t.Messages) == 0 {
		return nil
	}

	msgs := t.ids[topic]
	if msgs == nil || len(msgs) < idx {
		return nil
	}

	return msgs[idx]
}

// Close the tracker.
func (t *MockTracker) Close() {
}

func (t *MockTracker) FastMessageWithKey(topic string, value interface{}, key []byte) error {
	valueBuf, err := t.Encode(value)
	if err != nil {
		return err
	}

	t.Messages[topic] = append(t.Messages[topic], valueBuf)
	t.ids[topic] = append(t.ids[topic], key)
	return nil
}

func (t *MockTracker) SafeMessageWithKey(topic string, value interface{}, key []byte) error {
	return t.FastMessageWithKey(topic, value, key)
}

// FastMessage sends a message without waiting for confirmation.
func (t *MockTracker) FastMessage(topic string, value interface{}) error {
	return t.FastMessageWithKey(topic, value, nil)
}

// SafeMessage sends a message and waits for confirmation.
func (t *MockTracker) SafeMessage(topic string, value interface{}) error {
	return t.SafeMessageWithKey(topic, value, nil)
}

func (t *MockTracker) CheckHealth() error {
	return nil
}

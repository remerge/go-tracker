package tracker

import "sync"

// MockTracker is a tracker with in-memory storage used for testing.
type MockTracker struct {
	BaseTracker
	mapLock  sync.RWMutex
	Messages map[string][][]byte
	ids      map[string][][]byte
}

type MockTopicIterator struct {
	idxPosition int
	topic       string
	tracker     *MockTracker
}

//Next Return key, value, and false if there are no more messages
func (i *MockTopicIterator) Next() ([]byte, []byte, bool) {
	key := i.tracker.GetKey(i.topic, i.idxPosition)
	value := i.tracker.Get(i.topic, i.idxPosition)
	i.idxPosition++
	return key, value, len(i.tracker.Messages[i.topic]) > i.idxPosition
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

// Iterate returns you a topic iterator
func (t *MockTracker) Iterate(topic string) MockTopicIterator {
	return MockTopicIterator{topic: topic, tracker: t}
}

// Get a message from the mock broker.
func (t *MockTracker) Get(topic string, idx int) []byte {
	t.mapLock.RLock()
	defer t.mapLock.RUnlock()

	if len(t.Messages) == 0 {
		return nil
	}

	msgs := t.Messages[topic]
	if msgs == nil || len(msgs) <= idx {
		return nil
	}

	return msgs[idx]
}

// GetKey Gets a key from the mock broker.
func (t *MockTracker) GetKey(topic string, idx int) []byte {
	t.mapLock.RLock()
	defer t.mapLock.RUnlock()

	if len(t.Messages) == 0 {
		return nil
	}

	msgs := t.ids[topic]
	if msgs == nil || len(msgs) <= idx {
		return nil
	}

	return msgs[idx]
}

// Close the tracker.
func (t *MockTracker) Close() {
}

func (t *MockTracker) FastMessageWithKey(topic string, value interface{}, key []byte) error {
	t.mapLock.Lock()
	defer t.mapLock.Unlock()

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

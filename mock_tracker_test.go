package tracker

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMockTrackerBasic(t *testing.T) {
	topic := "topic"
	subject := NewMockTracker(&EventMetadata{})
	err := subject.FastMessage(topic, []byte("message1"))
	require.Nil(t, err)
	err = subject.FastMessage(topic, []byte("message2"))
	require.Nil(t, err)
	err = subject.FastMessageWithKey(topic, []byte("message3"), []byte("key3"))
	require.Nil(t, err)

	value := subject.Get(topic, 0)
	require.Equal(t, "message1", string(value))
	value = subject.Get(topic, 1)
	require.Equal(t, "message2", string(value))
	value = subject.Get(topic, 2)
	require.Equal(t, "message3", string(value))
	value = subject.Get(topic, 3)
	require.Nil(t, value)
}

func TestMockTrackerTopicIteration(t *testing.T) {
	topic := "topic"
	subject := NewMockTracker(&EventMetadata{})
	err := subject.FastMessage(topic, []byte("message1"))
	require.Nil(t, err)
	err = subject.FastMessage(topic, []byte("message2"))
	require.Nil(t, err)
	err = subject.FastMessageWithKey(topic, []byte("message3"), []byte("key3"))
	require.Nil(t, err)
	it := subject.Iterate(topic)
	// Check iteration
	key, value, canContinue := it.Next()
	require.Nil(t, key)
	require.Equal(t, "message1", string(value))
	require.True(t, canContinue)

	key, value, canContinue = it.Next()
	require.Nil(t, key)
	require.Equal(t, "message2", string(value))
	require.True(t, canContinue)

	key, value, canContinue = it.Next()
	require.Equal(t, "key3", string(key))
	require.Equal(t, "message3", string(value))
	require.False(t, canContinue)

	key, value, canContinue = it.Next()
	require.Nil(t, key)
	require.Nil(t, value)
	require.False(t, canContinue)
}

func TestNilMockTrackerBasic(t *testing.T) {
	metadata := EventMetadata{}
	topic := "topic"
	subject := NewMockTracker(&metadata)

	err := subject.FastMessage(topic, []byte(nil))
	require.Nil(t, err)
	err = subject.FastMessage(topic, []byte("message"))
	require.Nil(t, err)

	it := subject.Iterate(topic)

	key, value, canContinue := it.Next()
	require.Nil(t, key)
	require.Nil(t, value)
	require.True(t, canContinue)

	key, value, canContinue = it.Next()
	require.Nil(t, key)
	require.Equal(t, "message", string(value))
	require.False(t, canContinue)
}

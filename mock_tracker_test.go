package tracker

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMockTrackerBasic(t *testing.T) {
	metadata := EventMetadata{}
	topic := "topic"
	subject := NewMockTracker(&metadata)
	err := subject.FastMessage(topic, []byte("message1"))
	require.Nil(t, err)
	err = subject.FastMessage(topic, []byte("message2"))
	require.Nil(t, err)
	err = subject.FastMessageWithKey(topic, []byte("message3"), []byte("key3"))
	require.Nil(t, err)

	value := subject.Get(topic, 0)
	require.Equal(t, string(value), "message1")
	value = subject.Get(topic, 1)
	require.Equal(t, string(value), "message2")
	value = subject.Get(topic, 2)
	require.Equal(t, string(value), "message3")
	value = subject.Get(topic, 3)
	require.Nil(t, value)

	// Check iteration
	key, value := subject.Next(topic)
	require.Nil(t, key)
	require.Equal(t, string(value), "message1")
	key, value = subject.Next(topic)
	require.Nil(t, key)
	require.Equal(t, string(value), "message2")
	key, value = subject.Next(topic)
	require.Equal(t, string(key), "key3")
	require.Equal(t, string(value), "message3")
	key, value = subject.Next(topic)
	require.Nil(t, key)
	require.Nil(t, value)
}

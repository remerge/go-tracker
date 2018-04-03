package tracker

import (
	"runtime"
	"testing"

	"github.com/Shopify/sarama"
	metrics "github.com/rcrowley/go-metrics"
	"github.com/stretchr/testify/assert"
)

func newKafkaTrackerForTests(brokers []string,
	metadata *EventMetadata) (t *KafkaTracker, err error) {
	return NewKafkaTrackerConfig(KafkaTrackerConfig{
		Brokers:         brokers,
		Metadata:        metadata,
		MetricsRegistry: metrics.DefaultRegistry,
		Compression:     sarama.CompressionNone,
	})
}

func TestKafkaTrackerMemoryAllocation(t *testing.T) {
	//cue.Collect(cue.DEBUG, collector.Terminal{}.New())

	var memStart, memFinish runtime.MemStats
	runtime.ReadMemStats(&memStart)

	metadata := &EventMetadata{
		Service:     "tracker",
		Environment: "development",
		Cluster:     "local",
		Host:        "localhost",
		Release:     "master",
	}

	kt, err := newKafkaTrackerForTests([]string{"localhost:9092"}, metadata)
	assert.Nil(t, err)
	lt := NewLogTracker("log", metadata)
	messagesCount := 100

	for i := 0; i < messagesCount; i++ {
		err = kt.FastMessage("test", "message")
		assert.NoError(t, err)
		err = lt.FastMessage("test", "message")
		assert.NoError(t, err)
		err = kt.SafeMessage("test", "message")
		assert.NoError(t, err)
		err = lt.SafeMessage("test", "message")
		assert.NoError(t, err)
	}
	runtime.ReadMemStats(&memFinish)
	totalAllocations := memFinish.Mallocs - memStart.Mallocs
	const allowedAllocations = 19000
	t.Logf("Used %v from %v", totalAllocations, allowedAllocations)
	if totalAllocations > allowedAllocations {
		t.Errorf("Total number of allocations %v", totalAllocations)
	}
}

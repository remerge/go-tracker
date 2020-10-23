package tracker

import (
	"fmt"
	"time"

	"github.com/Shopify/sarama"
	"github.com/rcrowley/go-metrics"
	"github.com/remerge/cue"
)

// KafkaTrackerConfig is init configuration for KafkaTracker
type KafkaTrackerConfig struct {
	Brokers         []string
	Metadata        *EventMetadata
	MetricsRegistry metrics.Registry
	Compression     sarama.CompressionCodec
}

// KafkaTracker is a tracker that sends messages to Apache Kafka.
type KafkaTracker struct {
	BaseTracker
	metrics struct {
		registry      metrics.Registry
		fastErrorRate metrics.Meter
		safeErrorRate metrics.Meter
	}
	kafka struct {
		fast sarama.AsyncProducer
		safe sarama.SyncProducer
	}
}

var _ Tracker = (*KafkaTracker)(nil)

// NewKafkaTracker creates a new tracker connected to a kafka cluster.
func NewKafkaTracker(brokers []string,
	metadata *EventMetadata) (t *KafkaTracker, err error) {
	return NewKafkaTrackerConfig(KafkaTrackerConfig{
		Brokers:         brokers,
		Metadata:        metadata,
		MetricsRegistry: metrics.DefaultRegistry,
		Compression:     sarama.CompressionSnappy,
	})
}

// NewKafkaTrackerConfig accepts configuration and creates a new tracker
// connected to a kafka cluster.
func NewKafkaTrackerConfig(trackerConfig KafkaTrackerConfig) (t *KafkaTracker,
	err error) {
	log.WithValue("brokers", trackerConfig.Brokers).Info("starting tracker")

	t = &KafkaTracker{}
	t.Metadata = trackerConfig.Metadata
	t.metrics.registry = trackerConfig.MetricsRegistry

	// fast producer
	config := sarama.NewConfig()
	config.Version = sarama.V2_3_0_0
	config.ClientID = fmt.Sprintf(
		"tracker.fast-%s-%s-%s-%s",
		trackerConfig.Metadata.Environment,
		trackerConfig.Metadata.Cluster,
		trackerConfig.Metadata.Host,
		trackerConfig.Metadata.Service,
	)

	config.Producer.Return.Successes = false
	config.Producer.Return.Errors = true
	config.Producer.RequiredAcks = sarama.NoResponse
	config.Producer.Compression = trackerConfig.Compression
	config.ChannelBufferSize = 131072
	config.MetricRegistry = metrics.NewPrefixedChildRegistry(
		t.metrics.registry, "tracker,type=fast kafka_")
	t.metrics.fastErrorRate = metrics.GetOrRegisterMeter(
		"tracker,type=fast kafka_produce_error_rate",
		t.metrics.registry)

	t.kafka.fast, err = sarama.NewAsyncProducer(trackerConfig.Brokers, config)
	if err != nil {
		return nil, err
	}
	go func() {
		for fastErr := range t.kafka.fast.Errors() {
			t.metrics.fastErrorRate.Mark(1)
			if log.EnabledFor(cue.WARN) {
				log.WithFields(cue.Fields{
					"topic": fastErr.Msg.Topic,
				}).Warnf("send fast message failed", fastErr.Err)
			}
		}
	}()

	// safe producer
	config = sarama.NewConfig()
	config.Version = sarama.V2_3_0_0
	config.ClientID = fmt.Sprintf(
		"tracker.safe-%s-%s-%s-%s",
		trackerConfig.Metadata.Environment,
		trackerConfig.Metadata.Cluster,
		trackerConfig.Metadata.Host,
		trackerConfig.Metadata.Service,
	)

	config.Producer.Retry.Max = 10
	config.Producer.Retry.Backoff = 200 * time.Millisecond
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Compression = trackerConfig.Compression
	config.MetricRegistry = metrics.NewPrefixedChildRegistry(
		t.metrics.registry, "tracker,type=safe kafka_")

	t.kafka.safe, err = sarama.NewSyncProducer(trackerConfig.Brokers, config)
	if err != nil {
		_ = log.Error(t.kafka.fast.Close(), "failed to close fast producer")
		return nil, err
	}
	t.metrics.safeErrorRate = metrics.GetOrRegisterMeter(
		"tracker,type=safe kafka_produce_error_rate",
		t.metrics.registry)

	return t, nil
}

// Close the tracker.
func (t *KafkaTracker) Close() {
	// shutdown safe producer
	log.Info("closing safe producer")
	_ = log.Error(t.kafka.safe.Close(), "failed to close safe producer")

	// shutdwn fast producer
	log.Info("closing fast producer")
	_ = log.Error(t.kafka.fast.Close(), "failed to close fast producer")
}

// FastMessage sends a message without waiting for confirmation.
func (t *KafkaTracker) FastMessage(topic string, value interface{}) error {
	return t.FastMessageWithKey(topic, value, nil)
}

// FastMessageWithKey sends a message without waiting for confirmation.
func (t *KafkaTracker) FastMessageWithKey(topic string, value interface{}, key []byte) error {
	message, err := t.generateMessage(topic, "fast", value, key)
	if err != nil {
		return err
	}

	t.kafka.fast.Input() <- message

	return nil
}

// SafeMessage sends a message and waits for confirmation.
func (t *KafkaTracker) SafeMessage(topic string, value interface{}) error {
	return t.SafeMessageWithKey(topic, value, nil)
}

// SafeMessageWithKey sends a message and waits for confirmation.
func (t *KafkaTracker) SafeMessageWithKey(topic string, value interface{}, key []byte) error {
	message, err := t.generateMessage(topic, "safe", value, key)
	if err != nil {
		return err
	}

	_, _, err = t.kafka.safe.SendMessage(message)
	if err != nil {
		t.metrics.safeErrorRate.Mark(1)
	}

	return err
}

func (t *KafkaTracker) generateMessage(topic string, typ string, value interface{}, key []byte) (*sarama.ProducerMessage, error) {
	message := sarama.ProducerMessage{
		Topic: topic,
	}

	encodedValue, err := t.Encode(value)
	if err != nil {
		return nil, err
	}

	if log.EnabledFor(cue.DEBUG) {
		log.WithFields(cue.Fields{
			"topic": topic,
			"value": string(encodedValue),
			"key":   key,
		}).Debugf("sending %s message", typ)
	}

	if encodedValue != nil {
		message.Value = sarama.ByteEncoder(encodedValue)
	}
	if key != nil {
		message.Key = sarama.ByteEncoder(key)
	}

	return &message, nil
}

func (t *KafkaTracker) CheckHealth() (err error) {
	safeRate := t.metrics.safeErrorRate.Rate1()
	fastRate := t.metrics.fastErrorRate.Rate1()
	if safeRate > 0.05 || fastRate > 0.05 {
		return fmt.Errorf(`send errors: fast=%f safe=%f`, safeRate, fastRate)
	}
	return nil
}

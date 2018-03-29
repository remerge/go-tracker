package tracker

import (
	"fmt"

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

// NewKafkaTrackerForTests does not add compression
// since it is not working on wurstmeister
func NewKafkaTrackerForTests(brokers []string,
	metadata *EventMetadata) (t *KafkaTracker, err error) {
	return NewKafkaTrackerConfig(KafkaTrackerConfig{
		Brokers:         brokers,
		Metadata:        metadata,
		MetricsRegistry: metrics.DefaultRegistry,
		Compression:     sarama.CompressionNone,
	})
}

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
	config.Version = sarama.V0_10_0_0
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
			log.WithFields(cue.Fields{
				"topic": fastErr.Msg.Topic,
			}).Warnf("send fast message failed", fastErr.Err)
		}
	}()

	// safe producer
	config = sarama.NewConfig()
	config.Version = sarama.V0_10_0_0
	config.ClientID = fmt.Sprintf(
		"tracker.safe-%s-%s-%s-%s",
		trackerConfig.Metadata.Environment,
		trackerConfig.Metadata.Cluster,
		trackerConfig.Metadata.Host,
		trackerConfig.Metadata.Service,
	)

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
func (t *KafkaTracker) FastMessage(topic string, message interface{}) error {
	buf, err := t.Encode(message)
	if err != nil {
		return err
	}

	log.WithFields(cue.Fields{
		"topic":   topic,
		"message": string(buf),
	}).Debug("sending fast message")

	t.kafka.fast.Input() <- &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(buf),
	}

	return nil
}

// SafeMessage sends a message and waits for confirmation.
func (t *KafkaTracker) SafeMessage(topic string, message interface{}) error {
	buf, err := t.Encode(message)
	if err != nil {
		return err
	}

	log.WithFields(cue.Fields{
		"topic":   topic,
		"message": string(buf),
	}).Debug("sending safe message")

	_, _, err = t.kafka.safe.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(buf),
	})
	if err != nil {
		t.metrics.safeErrorRate.Mark(1)
	}

	return err
}

func (t *KafkaTracker) CheckHealth() (err error) {
	safeRate := t.metrics.safeErrorRate.Rate1()
	fastRate := t.metrics.fastErrorRate.Rate1()
	if safeRate > 0.05 || fastRate > 0.05 {
		return fmt.Errorf(`send errors: fast=%f safe=%f`, safeRate, fastRate)
	}
	return nil
}

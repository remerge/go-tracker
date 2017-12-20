package tracker

import (
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/rcrowley/go-metrics"
	"github.com/remerge/cue"
)

// KafkaTracker is a tracker that sends messages to Apache Kafka.
type KafkaTracker struct {
	BaseTracker
	metrics struct {
		registry metrics.Registry
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
func NewKafkaTracker(
	brokers []string,
	metadata *EventMetadata,
) (t *KafkaTracker, err error) {
	log.WithValue("brokers", brokers).Info("starting tracker")

	t = &KafkaTracker{}
	t.Metadata = metadata

	// TODO (Aleksandr Dorofeev): Add metrics registry to arguments
	t.metrics.registry = metrics.DefaultRegistry

	// fast producer
	config := sarama.NewConfig()
	config.ClientID = fmt.Sprintf(
		"tracker.fast-%s-%s-%s-%s",
		metadata.Environment,
		metadata.Cluster,
		metadata.Host,
		metadata.Service,
	)

	config.Producer.Return.Successes = false
	config.Producer.Return.Errors = true
	config.Producer.RequiredAcks = sarama.NoResponse
	config.Producer.Compression = sarama.CompressionSnappy
	config.ChannelBufferSize = 131072
	config.MetricRegistry = metrics.NewPrefixedChildRegistry(
		t.metrics.registry, "tracker,type=fast kafka_")
	t.metrics.fastErrorRate = metrics.GetOrRegisterMeter(
		"tracker,type=fast kafka_produce_error_rate",
		nil)

	t.kafka.fast, err = sarama.NewAsyncProducer(brokers, config)
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
	config.ClientID = fmt.Sprintf(
		"tracker.safe-%s-%s-%s-%s",
		metadata.Environment,
		metadata.Cluster,
		metadata.Host,
		metadata.Service,
	)

	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Compression = sarama.CompressionSnappy
	config.MetricRegistry = metrics.NewPrefixedChildRegistry(
		t.metrics.registry, "tracker,type=safe kafka_")

	t.kafka.safe, err = sarama.NewSyncProducer(brokers, config)
	if err != nil {
		_ = log.Error(t.kafka.fast.Close(), "failed to close fast producer")
		return nil, err
	}
	t.metrics.safeErrorRate = metrics.GetOrRegisterMeter(
		"tracker,type=safe kafka_produce_error_rate",
		nil)

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

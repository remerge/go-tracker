package tracker

import (
	"bytes"
	"fmt"
	"time"

	"github.com/Shopify/sarama"
	"github.com/beeker1121/goque"
)

// KafkaTracker is a tracker that sends messages to Apache Kafka.
type KafkaTracker struct {
	BaseTracker

	queue *goque.Queue
	quit  chan bool
	done  chan bool

	kafka struct {
		fast sarama.AsyncProducer
		safe sarama.SyncProducer
	}
}

var _ Tracker = (*KafkaTracker)(nil)

// NewKafkaTracker creates a new mock tracker for testing.
func NewKafkaTracker(brokers []string, metadata *EventMetadata) (t *KafkaTracker, err error) {
	t = &KafkaTracker{}
	t.Metadata = metadata

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
	config.Producer.Return.Errors = false
	config.Producer.RequiredAcks = sarama.NoResponse
	config.Producer.Compression = sarama.CompressionSnappy
	config.ChannelBufferSize = 131072

	t.kafka.fast, err = sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

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

	t.kafka.safe, err = sarama.NewSyncProducer(brokers, config)
	if err != nil {
		t.kafka.fast.Close()
		return nil, err
	}

	t.queue, err = goque.OpenQueue("cache/" + config.ClientID)
	if err != nil {
		t.kafka.fast.Close()
		t.kafka.safe.Close()
		return nil, err
	}

	t.quit = make(chan bool)
	t.done = make(chan bool)

	go t.start()

	return t, nil
}

// Close the tracker.
func (t *KafkaTracker) Close() {
	// shutdown background worker
	close(t.quit)
	<-t.done

	// shutdown safe producer
	t.kafka.safe.Close()

	// shutdwn fast producer
	t.kafka.fast.Close()

	// close access to disk queue
	t.queue.Close()
}

// FastMessage sends a message without waiting for confirmation.
func (t *KafkaTracker) FastMessage(topic string, message interface{}) error {
	buf, err := t.Encode(message)
	if err != nil {
		return err
	}

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

	return t.enqueue(topic, buf)
}

// fail-safe disk queue worker

var safeQueueDelim = []byte{0x0}

func (t *KafkaTracker) enqueue(topic string, value []byte) error {
	msg := append([]byte(topic), safeQueueDelim...)
	msg = append(msg, value...)
	_, err := t.queue.Enqueue(msg)
	return err
}

func (t *KafkaTracker) processSafeMessage(msg []byte) error {
	idx := bytes.Index(msg, safeQueueDelim)
	topic := string(msg[0:idx])
	value := msg[idx+1:]

	_, _, err := t.kafka.safe.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(value),
	})

	return err
}

func (t *KafkaTracker) start() {
	for {
		select {
		default:
			item, err := t.queue.Dequeue()
			if err == goque.ErrEmpty {
				time.Sleep(100 * time.Millisecond)
			} else if err != nil {
				panic(err)
			} else if err = t.processSafeMessage(item.Value); err != nil {
				sarama.Logger.Printf("failed to send safe message: %s\n", err.Error())
				t.queue.Enqueue(item.Value)
			}
		case <-t.quit:
			close(t.done)
			return
		}
	}
}

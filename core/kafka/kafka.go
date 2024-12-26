package kafka

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

type ProviderI interface {
	PublishToKafka(topic string, key, message []byte) error
	ReconnectWriter(topic string) error
	CloseWriter(topic string) error
	CloseAllWriters() error
}

type KafkaProvider struct {
	Host         string
	WriteTimeout time.Duration
	Config       *sarama.Config
	mutex        sync.RWMutex
	writers      map[string]sarama.AsyncProducer
}

var (
	BlobberMonitoringKafkaTopic = "blobber_monitoring"
	BlobberMonitoringKafka      = NewKafkaProvider("91.107.200.12:9092", "admin", "zus-operator", 1*time.Minute)
)

func NewKafkaProvider(host, username, password string, writeTimeout time.Duration) *KafkaProvider {
	log.Println("Initializing Kafka provider", zap.String("host", host))

	config := sarama.NewConfig()
	config.Net.SASL.Enable = true
	config.Net.SASL.User = username
	config.Net.SASL.Password = password
	config.Net.SASL.Mechanism = sarama.SASLTypePlaintext
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Net.MaxOpenRequests = 1
	config.Producer.Idempotent = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Metadata.AllowAutoTopicCreation = true
	config.Producer.MaxMessageBytes = 10 * 1024 * 1024

	return &KafkaProvider{
		Host:         host,
		WriteTimeout: writeTimeout,
		Config:       config,
		writers:      make(map[string]sarama.AsyncProducer),
	}
}

func (k *KafkaProvider) PublishToKafka(topic string, key, message []byte) error {
	k.mutex.RLock()
	writer, exists := k.writers[topic]
	k.mutex.RUnlock()

	if !exists {
		k.mutex.Lock()
		writer, exists = k.writers[topic]
		if !exists {
			writer = k.createKafkaWriter(topic)
			k.writers[topic] = writer
		}
		k.mutex.Unlock()
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.ByteEncoder(key),
		Value: sarama.ByteEncoder(message),
	}

	select {
	case writer.Input() <- msg:
		log.Println("Message published to Kafka", zap.String("topic", topic))
	case <-time.After(k.WriteTimeout):
		log.Println("Failed to publish to Kafka: timeout", zap.String("topic", topic))
		return fmt.Errorf("timeout publishing to Kafka topic %s", topic)
	}

	select {
	case <-writer.Successes():
		log.Println("Message published to Kafka", zap.String("topic", topic))
	case err := <-writer.Errors():
		log.Println("Failed to publish to Kafka", zap.String("topic", topic), zap.Error(err))
	}

	return nil
}

func PublishBlobberMonitoringLogsToKafka(key, message []byte) error {
	return BlobberMonitoringKafka.PublishToKafka(BlobberMonitoringKafkaTopic, key, message)
}

func (k *KafkaProvider) ReconnectWriter(topic string) error {
	k.mutex.Lock()
	defer k.mutex.Unlock()

	writer, exists := k.writers[topic]
	if !exists {
		return fmt.Errorf("no Kafka writer found for topic %v", topic)
	}

	if err := writer.Close(); err != nil {
		log.Println("Error closing Kafka writer", zap.Error(err))
		return fmt.Errorf("error closing Kafka connection for topic %v: %v", topic, err)
	}

	k.writers[topic] = k.createKafkaWriter(topic)
	return nil
}

func (k *KafkaProvider) CloseWriter(topic string) error {
	k.mutex.Lock()
	defer k.mutex.Unlock()

	writer, exists := k.writers[topic]
	if !exists {
		return fmt.Errorf("no Kafka writer found for topic %v", topic)
	}

	if err := writer.Close(); err != nil {
		log.Println("Error closing Kafka writer", zap.Error(err))
		return err
	}

	delete(k.writers, topic)
	return nil
}

func (k *KafkaProvider) CloseAllWriters() error {
	k.mutex.Lock()
	defer k.mutex.Unlock()

	for topic, writer := range k.writers {
		if err := writer.Close(); err != nil {
			log.Println("Error closing Kafka writer", zap.String("topic", topic), zap.Error(err))
		}
		delete(k.writers, topic)
	}
	return nil
}

func (k *KafkaProvider) createKafkaWriter(topic string) sarama.AsyncProducer {
	producer, err := sarama.NewAsyncProducer([]string{k.Host}, k.Config)
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}

	go func() {
		for err := range producer.Errors() {
			log.Printf("Kafka error: %v\n", err)
		}
	}()

	return producer
}

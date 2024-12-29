package kafka

import (
	"fmt"
	"github.com/0chain/common/core/logging"
	"github.com/0chain/gosdk/core/conf"
	"log"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

type BlobberMonitoring struct {
	ID           string `json:"id"`
	Operation    string `json:"operation"`
	BlobberId    string `json:"blobber_id"`
	AllocationId string `json:"allocation_id"`
	Size         int64  `json:"size"`
	TimeSpent    int64  `json:"time_spent"`
	Count        int    `json:"count"`
}

type ProviderI interface {
	PublishToKafka(topic string, key, message []byte) chan int64
	ReconnectWriter(topic string) error
	CloseWriter(topic string) error
	CloseAllWriters() error
}

type KafkaProvider struct {
	Host         string
	WriteTimeout time.Duration
	Config       *sarama.Config
	mutex        sync.RWMutex // Mutex for synchronizing access to writers map
}

// map of kafka writers for each topic
var writers map[string]sarama.AsyncProducer

func init() {
	writers = make(map[string]sarama.AsyncProducer)
}

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
	}
}

func PublishToKafka(key, message string) chan int64 {
	var (
		cfg                         = conf.GetConfig()
		BlobberMonitoringKafkaTopic = cfg.KafkaTopic
		BlobberMonitoringKafka      = NewKafkaProvider(cfg.KafkaHost, cfg.KafkaUsername, cfg.KafkaPassword, 1*time.Minute)
	)

	return BlobberMonitoringKafka.PublishToKafka(BlobberMonitoringKafkaTopic, key, message)
}

func (k *KafkaProvider) PublishToKafka(topic string, key, message string) chan int64 {
	k.mutex.RLock()
	writer, exists := writers[topic]
	k.mutex.RUnlock()
	res := make(chan int64)
	if !exists {
		k.mutex.Lock()
		writer, exists = writers[topic]
		if !exists {
			writer = k.createKafkaWriter(topic)
			writers[topic] = writer
		}
		k.mutex.Unlock()
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.ByteEncoder(key),
		Value: sarama.ByteEncoder(message),
	}

	fmt.Println("Publishing to kafka", zap.String("topic", topic), zap.String("key", key), zap.String("message", message))

	writer.Input() <- msg
	go func() {
		r := <-writer.Successes()
		res <- r.Offset
	}()
	return res
}

func (k *KafkaProvider) ReconnectWriter(topic string) error {
	k.mutex.Lock()
	defer k.mutex.Unlock()
	writer := writers[topic]
	if writer == nil {
		return fmt.Errorf("no kafka writer found for the topic %v", topic)
	}

	if err := writer.Close(); err != nil {
		logging.Logger.Error("error closing kafka connection", zap.String("topic", topic), zap.Error(err))
		return fmt.Errorf("error closing kafka connection for topic %v: %v", topic, err)
	}

	writers[topic] = k.createKafkaWriter(topic)
	return nil
}

func (k *KafkaProvider) CloseWriter(topic string) error {
	k.mutex.Lock()
	writer := writers[topic]
	k.mutex.Unlock()

	if writer == nil {
		return fmt.Errorf("no kafka writer found for the topic %v", topic)
	}

	if err := writer.Close(); err != nil {
		logging.Logger.Error("error closing kafka connection", zap.Error(err))
	}

	return nil
}

func (k *KafkaProvider) CloseAllWriters() error {
	k.mutex.Lock()
	defer k.mutex.Unlock()

	for topic, writer := range writers {
		if err := writer.Close(); err != nil {
			logging.Logger.Error("error closing kafka connection", zap.String("topic", topic), zap.Error(err))
		}
	}
	return nil
}

func (k *KafkaProvider) createKafkaWriter(topic string) sarama.AsyncProducer {
	producer, err := sarama.NewAsyncProducer([]string{k.Host}, k.Config)
	if err != nil {
		logging.Logger.Panic(fmt.Sprintf("Failed to start Sarama producer: %v", err))
	}

	go func() {
		for err := range producer.Errors() {
			fmt.Println("kafka - failed to write access log entry:", err)
			logging.Logger.Panic("kafka - failed to write access log entry:", zap.Error(err))
		}
	}()

	return producer
}

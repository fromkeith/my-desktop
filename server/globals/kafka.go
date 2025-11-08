package globals

import (
	"os"

	"github.com/segmentio/kafka-go"
)

var (
	kafkaConn *kafka.Conn
)

// auto closed.. general connection
func Kafka() *kafka.Conn {
	if kafkaConn != nil {
		return kafkaConn
	}
	var err error
	kafkaConn, err = kafka.Dial("tcp", os.Getenv("KAFKA_URI"))
	if err != nil {
		panic(err)
	}
	return kafkaConn
}

// A one off writer.. You are responsible for closing it
func KafkaWriter(topic string) *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP(os.Getenv("KAFKA_URI")),
		Topic:    "my-topic",
		Balancer: &kafka.LeastBytes{},
	}
}

// A one off reader.. You are responsible for closing it
func KafkaReader(topic string, partition int) *kafka.Reader {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{os.Getenv("KAFKA_URI")},
		Topic:     "my-topic",
		Partition: partition,
		MaxBytes:  10e6, // 10MB
	})
	// TODO: close on sigterm/sigint
	return r
}

// A one off reader.. You are responsible for closing it
func KafkaConsumerGroup(topic string, groupId string) *kafka.Reader {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{os.Getenv("KAFKA_URI")},
		Topic:    "my-topic",
		GroupID:  groupId,
		MaxBytes: 10e6, // 10MB
	})
	// TODO: close on sigterm/sigint
	return r
}

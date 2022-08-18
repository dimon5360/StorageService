package kafka

import (
	"app/main/utils"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type KafkaConfig struct {
	Host          string `json:"host"`
	ConsumerTopic string `json:"consumerTopic"`
	ProducerTopic string `json:"producerTopic"`
	GroupId       string `json:"group.id"`
}

type Handler struct {
	producer *kafka.Producer
	consumer *kafka.Consumer

	config *KafkaConfig
}

func InitKafkaHandler(jsonFileName string) *Handler {

	var handler Handler
	utils.ParseJsonConfig(jsonFileName, &handler.config)

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": handler.config.Host,
		"group.id":          handler.config.GroupId,
		"auto.offset.reset": "latest",
	})

	if err != nil {
		panic(err)
	}

	handler.consumer = consumer

	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": handler.config.Host,
	})
	if err != nil {
		panic(err)
	}

	handler.producer = producer

	return &handler
}

func (h *Handler) StartKafkaHandler(ConsumeCallback func(string, string)) {

	go h.StartKafkaConsumer(ConsumeCallback)
	go h.StartKafkaProducer()
}

func (h *Handler) StartKafkaConsumer(ConsumeCallback func(message string, topic string)) {

	err := h.consumer.Subscribe(h.config.ConsumerTopic, nil)
	if err != nil {
		panic(err)
	}
	log.Println("You have successfully connected to Kafka instance and subscribed the topic")

	for {
		msg, err := h.consumer.ReadMessage(10 * time.Millisecond)
		if err == nil {
			// log.Printf("Message on %s: %s\n", msg.TopicPartition, string(msg.Value))
			ConsumeCallback(msg.TopicPartition.String(), msg.String())
		} else if err.(kafka.Error).Code() != kafka.ErrTimedOut {
			log.Printf("Consumer error: %v (%v)\n", err, msg)
		}
	}
}

func (h *Handler) Produce(message string, topic string) {
	h.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          []byte(message),
	}, nil)
}

func (h *Handler) StartKafkaProducer() {

	go func() {
		for e := range h.producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					fmt.Printf("Delivery failed: %v\n", ev.TopicPartition)
				} else {
					fmt.Printf("Delivered message to %v\n", ev.TopicPartition)
				}
			case *kafka.Error:
				fmt.Fprintf(os.Stderr, "%% Error: %v\n", e)
			}
		}
	}()
}

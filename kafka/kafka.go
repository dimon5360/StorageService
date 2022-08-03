package kafka

import (
	"app/main/utils"
	"bufio"
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

func (h *Handler) StartKafkaHandler() {

	go h.KafkaConsumer()
	h.KafkaProducer()
}

func (h *Handler) KafkaConsumer() {

	err := h.consumer.Subscribe(h.config.ConsumerTopic, nil)
	if err != nil {
		panic(err)
	}
	log.Println("You have successfully connected to Kafka instance and subscribed the topic")

	for {
		msg, err := h.consumer.ReadMessage(10 * time.Millisecond)
		if err == nil {
			log.Printf("Message on %s: %s\n", msg.TopicPartition, string(msg.Value))
		} else if err.(kafka.Error).Code() != kafka.ErrTimedOut {
			log.Printf("Consumer error: %v (%v)\n", err, msg)
		}
	}
}

func (h *Handler) KafkaProducer() {

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

	reader := bufio.NewReader(os.Stdin)

	topic := h.config.ProducerTopic
	for {
		msg, _ := reader.ReadString('\n')
		// log.Printf("Produce message: %s", msg)
		h.producer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Value:          []byte(msg),
		}, nil)
	}
}

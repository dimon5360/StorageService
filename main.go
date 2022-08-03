package main

import (
	"app/main/kafka"
	"app/main/postgres"
	"log"
)

type AppConfig struct {
	postgresHandler *postgres.Handler
	kafkaHandler    *kafka.Handler
}

func main() {

	var config AppConfig
	log.Println("Start postgresql connection")

	config.postgresHandler = postgres.InitPostgresHandler("config/postgres.json", "config/create_tables.sql")
	config.kafkaHandler = kafka.InitKafkaHandler("config/kafka.json")

	config.kafkaHandler.StartKafkaHandler()
}

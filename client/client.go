package main

import (
	"app/main/postgres"
	"app/main/utils"
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type appConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

func main() {
	var app appConfig
	utils.ParseJsonConfig("../config/server.json", &app)

	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", app.Host, app.Port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := postgres.NewBarMapServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.CreateBar(ctx, &postgres.CreateBarRequest{
		Id:          "",
		Title:       "",
		Address:     "",
		Description: "",
		Drinks:      []*postgres.CreateDrinkRequest{},
	})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("All right %s\n", r.Address)
}

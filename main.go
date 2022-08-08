package main

import (
	"fmt"
	"log"
	"net"

	"app/main/postgres"
	"app/main/utils"

	"google.golang.org/grpc"
)

type appConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

func main() {
	log.Println("Start postgresql service")

	var app appConfig
	utils.ParseJsonConfig("config/server.json", &app)

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", app.Host, app.Port))
	if err != nil {
		panic(err)
	}
	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)
	postgres.RegisterBarMapServiceServer(grpcServer, postgres.NewServer("config/postgres.json", "config/create_tables.sql"))
	grpcServer.Serve(lis)
}

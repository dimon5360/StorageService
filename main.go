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

// build v.0.0.5 from 16.09.2022
const (
	BUILD = 5
	MINOR = 0
	MAJOR = 0
)

func main() {

	log.Printf("Database service v.%d.%d.%d.", MAJOR, MINOR, BUILD)

	var app appConfig
	utils.ParseJsonConfig("config/server.json", &app)

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", app.Host, app.Port))
	if err != nil {
		panic(err)
	}
	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)
	postgres.RegisterBarMapServiceServer(grpcServer, postgres.NewServer("config/postgres.json", "config/DrinksDatabaseDiagram.sql"))
	grpcServer.Serve(lis)
}

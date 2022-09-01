package postgres

import (
	"app/main/utils"

	"io/ioutil"
	"log"

	"github.com/jackc/pgx"
)

type PostgresConfig struct {
	Database string `json:"database"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type BarMapService struct {
	UnimplementedBarMapServiceServer

	handler *handler
}

type handler struct {
	config *PostgresConfig
	conn   *pgx.Conn
}

func NewServer(jsonFileName string, initScriptPath string) *BarMapService {

	handler := initPostgresHandler(jsonFileName, initScriptPath)
	return &BarMapService{
		handler: handler,
	}
}

func initPostgresHandler(jsonFileName string, initScriptPath string) *handler {

	var handler handler
	utils.ParseJsonConfig(jsonFileName, &handler.config)
	handler.conn = connectPostgres(*handler.config, initScriptPath)
	return &handler
}

func connectPostgres(conn_config PostgresConfig, initScriptPath string) *pgx.Conn {

	conn, err := pgx.Connect(pgx.ConnConfig{
		Host:     conn_config.Host,
		Port:     uint16(conn_config.Port),
		Database: conn_config.Database,
		User:     conn_config.User,
		Password: conn_config.Password,
	})

	if err != nil {
		panic(err)
	}

	c, err := ioutil.ReadFile(initScriptPath)
	if err != nil {
		panic(err)
	}

	sql := string(c)
	_, err = conn.Exec(sql)
	if err != nil {
		panic(err)
	}
	log.Println("Connection to database was succeed")

	return conn
}

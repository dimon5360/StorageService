package postgres

import (
	"app/main/utils"
	"io/ioutil"
	"log"
	"os"

	"github.com/jackc/pgx"
)

type PostgresConfig struct {
	Database string `json:"database"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type Handler struct {
	config *PostgresConfig
	conn   *pgx.Conn
}

func InitPostgresHandler(jsonFileName string, initScriptPath string) *Handler {

	var handler Handler
	utils.ParseJsonConfig(jsonFileName, &handler.config)
	handler.conn = handler.ConnectPostgres(*handler.config, initScriptPath)
	return &handler
}

func (h *Handler) ConnectPostgres(conn_config PostgresConfig, initScriptPath string) *pgx.Conn {

	conn, err := pgx.Connect(pgx.ConnConfig{
		Host:     conn_config.Host,
		Port:     uint16(conn_config.Port),
		Database: conn_config.Database,
		User:     conn_config.User,
		Password: conn_config.Password,
	})

	if err != nil {
		log.Printf("Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	log.Println("Connection to database was succeed")
	defer conn.Close()

	var responce string
	conn.QueryRow("select * from drinks;").Scan(&responce)

	c, err := ioutil.ReadFile(initScriptPath)
	if err != nil {
		panic(err)
	}

	sql := string(c)
	_, err = conn.Exec(sql)
	if err != nil {
		panic(err)
	}

	return conn
}

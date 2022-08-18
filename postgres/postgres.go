package postgres

import (
	"app/main/utils"
	context "context"
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

type barMapService struct {
	UnimplementedBarMapServiceServer

	handler *Handler
	// mu         sync.Mutex // protects routeNotes
}

func (s *barMapService) CreateBar(ctx context.Context, req *CreateBarRequest) (*Bar, error) {
	return &Bar{}, nil
}
func (s *barMapService) UpdateBar(ctx context.Context, req *UpdateBarRequest) (*Bar, error) {
	return &Bar{}, nil
}
func (s *barMapService) DeleteBar(ctx context.Context, req *DeleteBarRequest) (*DeleteBarResponse, error) {
	return &DeleteBarResponse{}, nil
}
func (s *barMapService) ListBar(ctx context.Context, req *ListBarsRequest) (*ListBarsResponse, error) {
	return &ListBarsResponse{}, nil
}
func (s *barMapService) GetBar(ctx context.Context, req *GetBarRequest) (*Bar, error) {
	return &Bar{}, nil
}
func (s *barMapService) CreateDrink(ctx context.Context, req *CreateDrinkRequest) (*Drink, error) {
	return &Drink{}, nil
}

func (s *barMapService) UpdateDrink(ctx context.Context, req *UpdateDrinkRequest) (*Drink, error) {
	return &Drink{}, nil
}
func (s *barMapService) DeleteDrink(ctx context.Context, req *DeleteDrinkRequest) (*DeleteDrinkResponse, error) {
	return &DeleteDrinkResponse{}, nil
}
func (s *barMapService) ListDrink(ctx context.Context, req *ListDrinksRequest) (*ListDrinksResponse, error) {
	return &ListDrinksResponse{}, nil
}
func (s *barMapService) GetDrink(ctx context.Context, req *GetDrinkRequest) (*Drink, error) {
	return &Drink{}, nil
}
func (s *barMapService) CreateIngredient(ctx context.Context, req *CreateIngredientRequest) (*Ingredient, error) {
	return &Ingredient{}, nil
}
func (s *barMapService) UpdateIngredient(ctx context.Context, req *UpdateIngredientRequest) (*Ingredient, error) {
	return &Ingredient{}, nil
}
func (s *barMapService) DeleteIngredient(ctx context.Context, req *DeleteIngredientRequest) (*DeleteIngredientResponse, error) {
	return &DeleteIngredientResponse{}, nil
}

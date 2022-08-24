package postgres

import (
	"app/main/utils"
	context "context"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PostgresConfig struct {
	Database string `json:"database"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type barMapService struct {
	UnimplementedBarMapServiceServer

	handler *handler
	// mu         sync.Mutex // protects routeNotes
}
type handler struct {
	config *PostgresConfig
	conn   *pgx.Conn
}

func NewServer(jsonFileName string, initScriptPath string) *barMapService {

	handler := initPostgresHandler(jsonFileName, initScriptPath)
	return &barMapService{
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

func (s *barMapService) CreateBar(ctx context.Context, req *CreateBarRequest) (*Bar, error) {

	var now = time.Now().Format("2006-01-02 15:04:05.000000")

	GetItemIds := func(Drinks []*CreateDrinkRequest) string {
		var ids string = "{"
		for i, drink := range Drinks {
			ids += drink.Id
			if i == len(Drinks)-1 {
				break
			}
			ids += ","
		}
		ids += "}"
		return ids
	}

	var sql string = fmt.Sprintf("insert into bars "+
		"(title, address, description, drinks_id, created_at, updated_at) "+
		"values ('%s', '%s', '%s', '%s', '%s', '%s') returning *;",
		req.Title, req.Address, req.Description, GetItemIds(req.Drinks), now, now)

	rows, err := s.handler.conn.Query(sql)
	if err != nil {
		log.Println("Failed sql transaction")
		return &Bar{}, err
	}

	var Id uint32
	var Title string
	var Address string
	var Description string
	var Drinks_Id pgtype.Int4Array
	var CreatedAt pgtype.Timestamp
	var UpdatedAt pgtype.Timestamp

	for rows.Next() {
		err := rows.Scan(&Id, &Title, &Address, &Description, &Drinks_Id, &CreatedAt, &UpdatedAt)
		if err != nil {
			panic(err)
		}
	}

	var drinks []*Drink
	Drinks_Id.AssignTo(&drinks)

	return &Bar{
		Id:          fmt.Sprintf("%d", Id),
		Title:       Title,
		Address:     Address,
		Description: Description,
		Drinks:      drinks,
		CreatedAt:   timestamppb.New(CreatedAt.Time),
		UpdatedAt:   timestamppb.New(UpdatedAt.Time),
	}, nil
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

	var sql string = fmt.Sprintf("select * from bars where id = %s;", req.Id)

	rows, err := s.handler.conn.Query(sql)
	if err != nil {
		log.Println("Failed sql transaction")
		return &Bar{}, err
	}

	var Id uint32
	var Title string
	var Address string
	var Description string
	var Drinks_Id pgtype.Int4Array
	var CreatedAt pgtype.Timestamp
	var UpdatedAt pgtype.Timestamp

	for rows.Next() {
		err := rows.Scan(&Id, &Title, &Address, &Description, &Drinks_Id, &CreatedAt, &UpdatedAt)
		if err != nil {
			panic(err)
		}
	}

	var drinks []*Drink
	Drinks_Id.AssignTo(&drinks)

	return &Bar{
		Id:          fmt.Sprintf("%d", Id),
		Title:       Title,
		Address:     Address,
		Description: Description,
		Drinks:      drinks,
		CreatedAt:   timestamppb.New(CreatedAt.Time),
		UpdatedAt:   timestamppb.New(UpdatedAt.Time),
	}, nil

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

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

// table bars model and SQL requests

type barModel struct {
	Id          string
	Title       string
	Address     string
	Description string
	Drinks_Id   pgtype.Int4Array
	CreatedAt   pgtype.Timestamp
	UpdatedAt   pgtype.Timestamp
}

func WrapBarResponse(rows *pgx.Rows, err error) (*Bar, error) {

	if err != nil {
		log.Println("Failed bars drinks sql transaction")
		return &Bar{}, err
	}

	var model barModel

	for rows.Next() {
		err := rows.Scan(&model.Id, &model.Title, &model.Address, &model.Description,
			&model.Drinks_Id, &model.CreatedAt, &model.UpdatedAt)
		if err != nil {
			log.Println("Failed bar sql response scan: ", err)
			return &Bar{}, err
		}
	}

	var drinks []*Drink
	model.Drinks_Id.AssignTo(&drinks)

	return &Bar{
		Id:          model.Id,
		Title:       model.Title,
		Address:     model.Address,
		Description: model.Description,
		Drinks:      drinks,
		CreatedAt:   timestamppb.New(model.CreatedAt.Time),
		UpdatedAt:   timestamppb.New(model.UpdatedAt.Time),
	}, nil
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

	return WrapBarResponse(s.handler.conn.Query(sql))
}

func (s *barMapService) UpdateBar(ctx context.Context, req *UpdateBarRequest) (*Bar, error) {
	var now = time.Now().Format("2006-01-02 15:04:05.000000")

	tx, err := s.handler.conn.BeginEx(ctx, nil)
	if err != nil {
		log.Println("Failed start bar update sql transaction: ", err)
		return &Bar{}, err
	}

	var sql string = "begin transation;"

	GetDrinksIds := func(Drinks []*UpdateDrinkRequest) string {
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

	sql += fmt.Sprintf("update bars set title = '%s', address = '%s', description = '%s', drinks_id = '%s', "+
		"updated_at = '%s' where id = %s returning *;",
		req.Title, req.Address, req.Description, GetDrinksIds(req.Drinks), now, req.Id)

	rows, err := tx.Query(fmt.Sprintf("update bars set title = '%s', address = '%s', description = '%s', drinks_id = '%s', "+
		"updated_at = '%s' where id = %s returning *;",
		req.Title, req.Address, req.Description, GetDrinksIds(req.Drinks), now, req.Id))

	if err != nil {
		log.Println("Failed bars drinks sql transaction: ", err)
		return &Bar{}, err
	}
	var model barModel

	for rows.Next() {
		err := rows.Scan(&model.Id, &model.Title, &model.Address, &model.Description, &model.Description, &model.Drinks_Id, &model.CreatedAt, &model.UpdatedAt)
		if err != nil {
			log.Println("Failed bar sql response scan: ", err)
			return &Bar{}, err
		}
	}

	var drinks []*Drink
	model.Drinks_Id.AssignTo(&drinks)

	for _, drink := range req.Drinks {

		_, err := tx.Query(fmt.Sprintf("update drinks set title = '%s', price = '%s', type = '%s', description = '%s', "+
			"bar_id = '%s', updated_at = '%s' where id = %s;",
			drink.Title, drink.Price, drink.DrinkType, req.Description, drink.BarId, now, drink.Id))

		if err != nil {
			log.Println("Failed update drinks sql transaction: ", err)
			return &Bar{}, err
		}

		for _, ingredient := range drink.Ingredients {
			_, err := tx.Exec(fmt.Sprintf("update ingredients set title = '%s', amount = '%s', drink_id = '%s', updated_at = '%s' where id = %s;",
				ingredient.Title, ingredient.Amount, ingredient.DrinkId, now, ingredient.Id))
			if err != nil {
				log.Println("Ingredient's update transaction failed: ", err)
				return &Bar{}, err
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		log.Println("Commiting bar's update transaction failed. Roolback performed: ", err)
		return &Bar{}, err
	}

	return &Bar{
		Id:          model.Id,
		Title:       model.Title,
		Address:     model.Address,
		Description: model.Description,
		Drinks:      drinks,
		CreatedAt:   timestamppb.New(model.CreatedAt.Time),
		UpdatedAt:   timestamppb.New(model.UpdatedAt.Time),
	}, nil
}

func (s *barMapService) DeleteBar(ctx context.Context, req *DeleteBarRequest) (*DeleteBarResponse, error) {
	sql := fmt.Sprintf("delete from bars WHERE id = %s;", req.Id)
	_, err := s.handler.conn.Exec(sql)
	if err != nil {
		log.Println("Deleting drink failed")
		return &DeleteBarResponse{}, err
	}
	return &DeleteBarResponse{}, nil
}

func (s *barMapService) ListBar(ctx context.Context, req *ListBarsRequest) (*ListBarsResponse, error) {
	return &ListBarsResponse{}, nil
}

func (s *barMapService) GetBar(ctx context.Context, req *GetBarRequest) (*Bar, error) {

	var sql string = fmt.Sprintf("select * from bars where id = %s;", req.Id)
	return WrapBarResponse(s.handler.conn.Query(sql))
}

// table drinks model and SQL requests

type drinkModel struct {
	Id             string
	Title          string
	Price          string
	Type           DrinkType
	Description    string
	BarId          string
	Ingredients_Id pgtype.Int4Array
	CreatedAt      pgtype.Timestamp
	UpdatedAt      pgtype.Timestamp
}

func WrapDrinkResponse(rows *pgx.Rows, err error) (*Drink, error) {

	if err != nil {
		log.Println("Failed sql transaction")
		return &Drink{}, err
	}

	var model drinkModel

	for rows.Next() {
		err := rows.Scan(&model.Id, &model.Title, &model.Price, &model.Type, &model.Description,
			&model.BarId, &model.Ingredients_Id, &model.CreatedAt, &model.UpdatedAt)
		if err != nil {
			log.Println("Failed drink sql response scan: ", err)
			return &Drink{}, err
		}
	}

	var ingredients []*Ingredient
	model.Ingredients_Id.AssignTo(&ingredients)

	return &Drink{
		Id:          model.Id,
		Title:       model.Title,
		Price:       model.Price,
		DrinkType:   model.Type,
		Description: model.Description,
		BarId:       model.BarId,
		Ingredients: ingredients,
		CreatedAt:   timestamppb.New(model.CreatedAt.Time),
		UpdatedAt:   timestamppb.New(model.UpdatedAt.Time),
	}, nil
}

func (s *barMapService) CreateDrink(ctx context.Context, req *CreateDrinkRequest) (*Drink, error) {
	var now = time.Now().Format("2006-01-02 15:04:05.000000")

	CreateIngrediendsScript := func(Ingredients []*CreateIngredientRequest, timestamp string) string {
		var sql string = "insert into ingredients " +
			"(title, amount, drink_id, created_at, updated_at) values "

		for i, ingrediemt := range Ingredients {
			sql += fmt.Sprintf("('%s', '%s', (select drinks.id from drinks), '%s', '%s')",
				ingrediemt.Title, ingrediemt.Amount, timestamp, timestamp)
			if i == len(Ingredients)-1 {
				break
			}
			sql += ", "
		}
		return sql
	}

	var sql string = fmt.Sprintf("insert into drinks "+
		"(title, price, type, description, bar_id, ingredients_id, created_at, updated_at) "+
		"values ('%s', '%s', %d, '%s','%s', '%s', '%s', '%s') returning *;",
		req.Title, req.Price, req.DrinkType, req.Description, req.BarId, CreateIngrediendsScript(req.Ingredients, now), now, now)

	return WrapDrinkResponse(s.handler.conn.Query(sql))
}

func (s *barMapService) UpdateDrink(ctx context.Context, req *UpdateDrinkRequest) (*Drink, error) {
	var now = time.Now().Format("2006-01-02 15:04:05.000000")

	tx, err := s.handler.conn.BeginEx(ctx, nil)
	if err != nil {
		log.Println("Failed drink bar update sql transaction: ", err)
		return &Drink{}, err
	}

	rows, err := tx.Query(fmt.Sprintf("update drinks set title = '%s', price = '%s', type = '%s', description = '%s', "+
		"bar_id = '%s', updated_at = '%s' where id = %s returning *;",
		req.Title, req.Price, req.DrinkType, req.Description, req.BarId, now, req.Id))

	if err != nil {
		log.Println("Failed sql transaction")
		return &Drink{}, err
	}
	var model drinkModel

	for rows.Next() {
		err := rows.Scan(&model.Id, &model.Title, &model.Price, &model.Type, &model.Description,
			&model.BarId, &model.Ingredients_Id, &model.CreatedAt, &model.UpdatedAt)
		if err != nil {
			return &Drink{}, err
		}
	}

	var ingredients []*Ingredient
	model.Ingredients_Id.AssignTo(&ingredients)

	for _, ingredient := range req.Ingredients {
		_, err := tx.Exec(fmt.Sprintf("update ingredients set title = '%s', amount = '%s', drink_id = '%s', updated_at = '%s' where id = %s;",
			ingredient.Title, ingredient.Amount, ingredient.DrinkId, now, ingredient.Id))
		if err != nil {
			log.Println("Ingredient's update transaction failed")
			return &Drink{}, err
		}
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		log.Println("Commiting drink's update transaction failed. Roolback performed.")
		return &Drink{}, err
	}

	return &Drink{
		Id:          model.Id,
		Title:       model.Title,
		Price:       model.Price,
		DrinkType:   model.Type,
		Description: model.Description,
		BarId:       model.BarId,
		Ingredients: ingredients,
		CreatedAt:   timestamppb.New(model.CreatedAt.Time),
		UpdatedAt:   timestamppb.New(model.UpdatedAt.Time),
	}, nil
}

func (s *barMapService) DeleteDrink(ctx context.Context, req *DeleteDrinkRequest) (*DeleteDrinkResponse, error) {

	sql := fmt.Sprintf("delete from drinks WHERE id = %s;", req.Id)
	_, err := s.handler.conn.Exec(sql)
	if err != nil {
		log.Println("Deleting drink failed")
		return &DeleteDrinkResponse{}, err
	}
	return &DeleteDrinkResponse{}, nil
}

func (s *barMapService) ListDrink(ctx context.Context, req *ListDrinksRequest) (*ListDrinksResponse, error) {
	return &ListDrinksResponse{}, nil
}

func (s *barMapService) GetDrink(ctx context.Context, req *GetDrinkRequest) (*Drink, error) {
	var sql string = fmt.Sprintf("select * from drinks where id = %s;", req.Id)
	return WrapDrinkResponse(s.handler.conn.Query(sql))
}

// table drinks model and SQL requests
type ingredientModel struct {
	Id        string
	Title     string
	Amount    string
	DrinkId   string
	CreatedAt pgtype.Timestamp
	UpdatedAt pgtype.Timestamp
}

func WrapIngredientResponse(rows *pgx.Rows, err error) (*Ingredient, error) {

	if err != nil {
		log.Println("Failed sql transaction")
		return &Ingredient{}, err
	}

	var model ingredientModel

	for rows.Next() {
		err := rows.Scan(&model.Id, &model.Title, &model.Amount, &model.DrinkId, &model.CreatedAt, &model.UpdatedAt)
		if err != nil {
			log.Println("Failed ingredient sql response scan: ", err)
			return &Ingredient{}, err
		}
	}

	return &Ingredient{
		Id:        model.Id,
		Title:     model.Title,
		Amount:    model.Amount,
		DrinkId:   model.DrinkId,
		CreatedAt: timestamppb.New(model.CreatedAt.Time),
		UpdatedAt: timestamppb.New(model.UpdatedAt.Time),
	}, nil
}
func (s *barMapService) CreateIngredient(ctx context.Context, req *CreateIngredientRequest) (*Ingredient, error) {
	var now = time.Now().Format("2006-01-02 15:04:05.000000")

	var sql string = fmt.Sprintf("insert into ingredients "+
		"(title, amount, drink_id, created_at, updated_at) "+
		"values ('%s', '%s', '%s', '%s', '%s') returning *;",
		req.Title, req.Amount, req.DrinkId, now, now)

	return WrapIngredientResponse(s.handler.conn.Query(sql))
}

func (s *barMapService) UpdateIngredient(ctx context.Context, req *UpdateIngredientRequest) (*Ingredient, error) {
	var now = time.Now().Format("2006-01-02 15:04:05.000000")

	tx, err := s.handler.conn.BeginEx(ctx, nil)
	if err != nil {
		log.Println("Failed ingredient bar update sql transaction: ", err)
		return &Ingredient{}, err
	}

	var sql = fmt.Sprintf("update ingredients set title = '%s', amount = '%s', drink_id = '%s', updated_at = '%s' where id = %s returning *;",
		req.Title, req.Amount, req.DrinkId, now, req.Id)
	ingredient, err := WrapIngredientResponse(tx.Query(sql))

	if err != nil {
		log.Println("Updating ingredient failed: ", err)
		return &Ingredient{}, err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		log.Println("Commiting Ingredient's update transaction failed. Roolback performed: ", err)
		return &Ingredient{}, err
	}

	return ingredient, nil
}

func (s *barMapService) DeleteIngredient(ctx context.Context, req *DeleteIngredientRequest) (*DeleteIngredientResponse, error) {
	sql := fmt.Sprintf("delete from ingredients WHERE id = %s;", req.Id)
	_, err := s.handler.conn.Exec(sql)
	if err != nil {
		log.Println("Deleting ingredient failed")
		return &DeleteIngredientResponse{}, err
	}
	return &DeleteIngredientResponse{}, nil
}

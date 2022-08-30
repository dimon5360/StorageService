package postgres

import (
	context "context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// table drinks model and SQL requests
type drinkModel struct {
	Id             int32
	Title          string
	Price          int32
	Type           DrinkType
	Description    string
	BarId          int32
	Ingredients_Id pgtype.Int4Array
	CreatedAt      pgtype.Timestamp
	UpdatedAt      pgtype.Timestamp
}

const UNDEFINED_ID int32 = -1

func WrapDrinkResponse(rows *pgx.Rows, err error) (*Drink, error) {

	if err != nil {
		log.Println("Failed sql transaction", err)
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

	for _, idx := range model.Ingredients_Id.Elements {
		log.Println(idx.Int)
	}

	return &Drink{
		Id:          fmt.Sprintf("%d", model.Id),
		Title:       model.Title,
		Price:       fmt.Sprintf("%d", model.Price),
		DrinkType:   model.Type,
		Description: model.Description,
		BarId:       fmt.Sprintf("%d", model.BarId),
		Ingredients: ingredients, // @TODO: fill ingredients
		CreatedAt:   timestamppb.New(model.CreatedAt.Time),
		UpdatedAt:   timestamppb.New(model.UpdatedAt.Time),
	}, nil
}

func (s *BarMapService) CreateDrink(ctx context.Context, req *CreateDrinkRequest) (*Drink, error) {
	var now = time.Now().Format("2006-01-02 15:04:05.000000")

	insertDrinkScript := fmt.Sprintf("with ids as (INSERT INTO drinks (title, price, type, description, bar_id, ingredients_id, created_at, updated_at) "+
		" VALUES ('%s', '%s', %d, '%s', '%s', '%s', '%s', '%s') returning *)",
		req.Title, req.Price, req.DrinkType, req.Description, req.BarId, "{}", now, now)

	CreateIngrediendsScript := func(Ingredients []*CreateIngredientRequest, timestamp string) string {
		var tmp string = "INSERT INTO ingredients " +
			"(title, amount, drink_id, created_at, updated_at) VALUES "

		for i, ingrediemt := range Ingredients {
			tmp += fmt.Sprintf("('%s', '%s', (select ids.id from ids), '%s', '%s')",
				ingrediemt.Title, ingrediemt.Amount, timestamp, timestamp)
			if i == len(Ingredients)-1 {
				tmp += "returning id, drink_id;"
				break
			}
			tmp += ", "
		}
		return tmp
	}

	rows, err := s.handler.conn.Query(insertDrinkScript + CreateIngrediendsScript(req.Ingredients, now))
	if err != nil {
		log.Println("Failed ingredient sql response scan: ", err)
		return &Drink{}, err
	}

	IngrediendsIdsParseScript := func(rows *pgx.Rows) (string, int32) {
		var IngredientId, DrinkId int32
		var IngredientIds = "ARRAY["
		for rows.Next() {
			err := rows.Scan(&IngredientId, &DrinkId)
			if err != nil {
				log.Println("Failed ingredient sql response scan: ", err)
				return "", UNDEFINED_ID
			}
			IngredientIds += fmt.Sprintf("%d", IngredientId)
			IngredientIds += ","
		}
		return strings.TrimSuffix(IngredientIds, ",") + "]", DrinkId
	}

	val, id := IngrediendsIdsParseScript(rows)
	if len(val) == 0 || id == UNDEFINED_ID {
		log.Printf("Results parse failed. Return empty object")
		return &Drink{}, err
	}

	var sql = fmt.Sprintf("update drinks set ingredients_id = array_cat(ingredients_id, %s), updated_at = '%s' where id = %d;", val, now, id)
	_, err = s.handler.conn.Exec(sql)
	if err != nil {
		log.Println("Failed update drink sql transaction: ", err)
		return &Drink{}, err
	}
	// return &Drink{}, err
	return WrapDrinkResponse(s.handler.conn.Query(fmt.Sprintf("select * from drinks where id = %d;", id)))
}

func (s *BarMapService) UpdateDrink(ctx context.Context, req *UpdateDrinkRequest) (*Drink, error) {
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
		Id:          fmt.Sprintf("%d", model.Id),
		Title:       model.Title,
		Price:       fmt.Sprintf("%d", model.Price),
		DrinkType:   model.Type,
		Description: model.Description,
		BarId:       fmt.Sprintf("%d", model.BarId),
		CreatedAt:   timestamppb.New(model.CreatedAt.Time),
		UpdatedAt:   timestamppb.New(model.UpdatedAt.Time),
	}, nil
}

func (s *BarMapService) DeleteDrink(ctx context.Context, req *DeleteDrinkRequest) (*DeleteDrinkResponse, error) {

	sql := fmt.Sprintf("delete from drinks WHERE id = %s;", req.Id)
	_, err := s.handler.conn.Exec(sql)
	if err != nil {
		log.Println("Deleting drink failed")
		return &DeleteDrinkResponse{}, err
	}
	return &DeleteDrinkResponse{}, nil
}

func (s *BarMapService) ListDrink(ctx context.Context, req *ListDrinksRequest) (*ListDrinksResponse, error) {
	return &ListDrinksResponse{}, nil
}

func (s *BarMapService) GetDrink(ctx context.Context, req *GetDrinkRequest) (*Drink, error) {
	var sql string = fmt.Sprintf("select * from drinks where id = %s;", req.Id)
	return WrapDrinkResponse(s.handler.conn.Query(sql))
}

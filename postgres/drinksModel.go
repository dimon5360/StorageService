package postgres

import (
	context "context"
	"fmt"
	"log"
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
	Ingredients_Id pgtype.Int8Array
	CreatedAt      pgtype.Timestamp
	UpdatedAt      pgtype.Timestamp
}

const (
	UNDEFINED_ID int32 = -1
	DEBUG_OUTPUT bool  = false
)

// sql response wrapper
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
	for _, idx := range model.Ingredients_Id.Elements {
		ingredients = append(ingredients, &Ingredient{
			Id: fmt.Sprintf("%d", idx.Int),
		})
	}

	return &Drink{
		Id:          fmt.Sprintf("%d", model.Id),
		Title:       model.Title,
		Price:       fmt.Sprintf("%d", model.Price),
		DrinkType:   model.Type,
		Description: model.Description,
		BarId:       fmt.Sprintf("%d", model.BarId),
		Ingredients: ingredients,
		CreatedAt:   timestamppb.New(model.CreatedAt.Time),
		UpdatedAt:   timestamppb.New(model.UpdatedAt.Time),
	}, nil
}

// gRPC creating drink request handler
// established, remove comment later
func (s *BarMapService) CreateDrink(ctx context.Context, req *CreateDrinkRequest) (*Drink, error) {

	var now = time.Now().Format("2006-01-02 15:04:05.000000")

	insertDrinkScript := "begin transaction;\n"
	insertDrinkScript += fmt.Sprintf("WITH created_drink AS (INSERT INTO drinks (title, price, type, description, bar_id, ingredients_id, created_at, updated_at) "+
		"VALUES ('%s', '%s', %d, '%s', '%s', '%s', '%s', '%s') returning *)\n",
		req.Title, req.Price, req.DrinkType, req.Description, req.BarId, "{}", now, now)

	CreateIngrediendsScript := func(Ingredients []*CreateIngredientRequest, timestamp string) string {
		var tmp string = "INSERT INTO ingredients " +
			"(title, amount, drink_id, created_at, updated_at) VALUES "

		for i, ingrediemt := range Ingredients {
			tmp += fmt.Sprintf("('%s', '%s', (select created_drink.id from created_drink), '%s', '%s')",
				ingrediemt.Title, ingrediemt.Amount, timestamp, timestamp)
			if i == len(Ingredients)-1 {
				tmp += " returning *;\n"
				break
			}
			tmp += ", "
		}
		return tmp
	}

	insertDrinkScript += CreateIngrediendsScript(req.Ingredients, now)
	insertDrinkScript += "commit;\n"

	_, err := s.handler.conn.Exec(insertDrinkScript)
	if err != nil {
		return &Drink{}, err
	}

	insertDrinkScript = "begin;\n"
	insertDrinkScript += fmt.Sprintf("update drinks set ingredients_id = array_cat(ingredients_id, "+
		"array(select ingredients.id from ingredients where drink_id = (select drinks.id from drinks where title = '%s' AND bar_id = '%s'))), updated_at = '%s' "+
		"where id = (select drinks.id from drinks where title = '%s' AND bar_id = '%s');\n", req.Title, req.BarId, now, req.Title, req.BarId)

	insertDrinkScript += fmt.Sprintf("update bars set drinks_id = array_append(drinks_id, (select drinks.id from drinks where title = '%s' AND bar_id = '%s')), "+
		"updated_at = '%s' where id = '%s';\n", req.Title, req.BarId, now, req.BarId)
	insertDrinkScript += "commit;\n"

	_, err = s.handler.conn.Exec(insertDrinkScript)
	if err != nil {
		return &Drink{}, err
	}
	return WrapDrinkResponse(s.handler.conn.Query(fmt.Sprintf("select * from drinks where title = '%s' AND bar_id = '%s';", req.Title, req.BarId)))
}

// / gRPC updating drink request handler
// / established, remove comment later
func (s *BarMapService) UpdateDrink(ctx context.Context, req *UpdateDrinkRequest) (*Drink, error) {
	var now = time.Now().Format("2006-01-02 15:04:05.000000")

	var updateDrinkScript string = fmt.Sprintf("update drinks set title = '%s', price = '%s', type = %d, description = '%s', "+
		"updated_at = '%s' where id = %s returning *;",
		req.Title, req.Price, req.DrinkType, req.Description, now, req.Id)

	return WrapDrinkResponse(s.handler.conn.Query(updateDrinkScript))
}

// / gRPC deleting drink request handler
// / established, remove comment later
func (s *BarMapService) DeleteDrink(ctx context.Context, req *DeleteDrinkRequest) (*DeleteDrinkResponse, error) {

	var sql string = "begin;\n"
	sql += fmt.Sprintf("delete from drinks WHERE id = %s;\n", req.Id)
	sql += fmt.Sprintf("update bars set drinks_id = array_remove(drinks_id, %s) WHERE %s = ANY(drinks_id);\n", req.Id, req.Id)
	sql += "commit;"

	_, err := s.handler.conn.Exec(sql)
	if err != nil {
		log.Println("Deleting drink failed")
		return &DeleteDrinkResponse{}, err
	}
	return &DeleteDrinkResponse{}, nil
}

// / gRPC geting drinks' list request handler
// #TODO: need to implement
func (s *BarMapService) ListDrink(ctx context.Context, req *ListDrinksRequest) (*ListDrinksResponse, error) {
	return &ListDrinksResponse{}, nil
}

// / gRPC getting drink request handler
// / established, remove comment later
func (s *BarMapService) GetDrink(ctx context.Context, req *GetDrinkRequest) (*Drink, error) {

	var selectDrink string = fmt.Sprintf("select * from drinks where id = %s;", req.Id)

	var selectIngredients = fmt.Sprintf("select * from ingredients where drink_id = %s;", req.Id)

	var rows, err = s.handler.conn.Query(selectIngredients)
	if err != nil {
		log.Println("Getting drink's ingredients failed")
		return &Drink{}, err
	}

	var ingredients []*Ingredient

	var im IngredientModel
	for rows.Next() {
		err := rows.Scan(&im.Id, &im.Title, &im.Amount, &im.DrinkId, &im.CreatedAt, &im.UpdatedAt)
		if err != nil {
			log.Println("Failed ingredient sql response scan: ", err)
			return &Drink{}, err
		}

		ingredients = append(ingredients, &Ingredient{
			Id:        fmt.Sprintf("%d", im.Id),
			Title:     im.Title,
			Amount:    fmt.Sprintf("%d", im.Amount),
			DrinkId:   fmt.Sprintf("%d", im.DrinkId),
			CreatedAt: timestamppb.New(im.CreatedAt.Time),
			UpdatedAt: timestamppb.New(im.UpdatedAt.Time),
		})
	}

	rows, err = s.handler.conn.Query(selectDrink)
	if err != nil {
		log.Println("Getting drink failed")
		return &Drink{}, err
	}

	var dm drinkModel

	for rows.Next() {
		err := rows.Scan(&dm.Id, &dm.Title, &dm.Price, &dm.Type, &dm.Description,
			&dm.BarId, nil, &dm.CreatedAt, &dm.UpdatedAt)
		if err != nil {
			log.Println("Failed drink sql response scan: ", err)
			return &Drink{}, err
		}
	}

	return &Drink{
		Id:          fmt.Sprintf("%d", dm.Id),
		Title:       dm.Title,
		Price:       fmt.Sprintf("%d", dm.Price),
		DrinkType:   dm.Type,
		Description: dm.Description,
		BarId:       fmt.Sprintf("%d", dm.BarId),
		Ingredients: ingredients,
		CreatedAt:   timestamppb.New(dm.CreatedAt.Time),
		UpdatedAt:   timestamppb.New(dm.UpdatedAt.Time),
	}, nil
}

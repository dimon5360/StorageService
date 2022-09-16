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
type IngredientModel struct {
	Id        int32
	Title     string
	Amount    int32
	DrinkId   int32
	CreatedAt pgtype.Timestamp
	UpdatedAt pgtype.Timestamp
}

func WrapIngredientResponse(rows *pgx.Rows, err error) (*Ingredient, error) {

	if err != nil {
		log.Println("Failed sql transaction")
		return &Ingredient{}, err
	}

	var model IngredientModel

	for rows.Next() {
		err := rows.Scan(&model.Id, &model.Title, &model.Amount, &model.DrinkId, &model.CreatedAt, &model.UpdatedAt)
		if err != nil {
			log.Println("Failed ingredient sql response scan: ", err)
			return &Ingredient{}, err
		}
	}

	return &Ingredient{
		Id:        fmt.Sprintf("%d", model.Id),
		Title:     model.Title,
		Amount:    fmt.Sprintf("%d", model.Amount),
		DrinkId:   fmt.Sprintf("%d", model.DrinkId),
		CreatedAt: timestamppb.New(model.CreatedAt.Time),
		UpdatedAt: timestamppb.New(model.UpdatedAt.Time),
	}, nil
}

/// gRPC creating ingredient request handler
// #TODO: need to test
func (s *BarMapService) CreateIngredient(ctx context.Context, req *CreateIngredientRequest) (*Ingredient, error) {
	var now = time.Now().Format("2006-01-02 15:04:05.000000")

	insertIngredientScript := "begin;\n"
	insertIngredientScript += fmt.Sprintf("WITH ingredient_id AS (insert into ingredients "+
		"(title, amount, drink_id, created_at, updated_at) "+
		"values ('%s', '%s', '%s', '%s', '%s') returning *);",
		req.Title, req.Amount, req.DrinkId, now, now)
	insertIngredientScript += fmt.Sprintf("update drinks set ingredients_id = array_append(ingredients_id, "+
		"(select ingredient_id.id from ingredient_id) where id = '%s');", req.DrinkId)
	insertIngredientScript += "commit;\n"

	_, err := s.handler.conn.Exec(insertIngredientScript)
	if err != nil {
		return &Ingredient{}, err
	}
	return WrapIngredientResponse(s.handler.conn.Query(fmt.Sprintf("select * from ingredients where title = '%s' AND drink_id = '%s';", req.Title, req.DrinkId)))
}

/// gRPC updating ingredient request handler
// #TODO: need to test
func (s *BarMapService) UpdateIngredient(ctx context.Context, req *UpdateIngredientRequest) (*Ingredient, error) {
	var now = time.Now().Format("2006-01-02 15:04:05.000000")

	var sql = fmt.Sprintf("update ingredients set title = '%s', amount = '%s', updated_at = '%s' where id = %s returning *;",
		req.Title, req.Amount, now, req.Id)
	return WrapIngredientResponse(s.handler.conn.Query(sql))
}

/// gRPC deleting ingredient request handler
// #TODO: need to test
func (s *BarMapService) DeleteIngredient(ctx context.Context, req *DeleteIngredientRequest) (*DeleteIngredientResponse, error) {
	var sql string = "begin;\n"
	sql += fmt.Sprintf("delete from ingredients WHERE id = %s;", req.Id)
	sql += fmt.Sprintf("update drinks set ingredients_id = array_remove(ingredients_id, %s) WHERE %s = ANY(ingredients_id);\n", req.Id, req.Id)
	sql += "commit;"

	_, err := s.handler.conn.Exec(sql)
	if err != nil {
		log.Println("Deleting ingredient failed")
		return &DeleteIngredientResponse{}, err
	}
	return &DeleteIngredientResponse{}, nil
}

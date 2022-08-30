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
func (s *BarMapService) CreateIngredient(ctx context.Context, req *CreateIngredientRequest) (*Ingredient, error) {
	var now = time.Now().Format("2006-01-02 15:04:05.000000")

	var sql string = fmt.Sprintf("insert into ingredients "+
		"(title, amount, drink_id, created_at, updated_at) "+
		"values ('%s', '%s', '%s', '%s', '%s') returning *;",
		req.Title, req.Amount, req.DrinkId, now, now)

	return WrapIngredientResponse(s.handler.conn.Query(sql))
}

func (s *BarMapService) UpdateIngredient(ctx context.Context, req *UpdateIngredientRequest) (*Ingredient, error) {
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

func (s *BarMapService) DeleteIngredient(ctx context.Context, req *DeleteIngredientRequest) (*DeleteIngredientResponse, error) {
	sql := fmt.Sprintf("delete from ingredients WHERE id = %s;", req.Id)
	_, err := s.handler.conn.Exec(sql)
	if err != nil {
		log.Println("Deleting ingredient failed")
		return &DeleteIngredientResponse{}, err
	}
	return &DeleteIngredientResponse{}, nil
}

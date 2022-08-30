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

// table bars model and SQL requests

type barModel struct {
	Id          int32
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
			log.Println("Failed bar sql response scan:", err)
			return &Bar{}, err
		}
	}

	var drinks []*Drink
	model.Drinks_Id.AssignTo(&drinks)

	return &Bar{
		Id:          fmt.Sprintf("%d", model.Id),
		Title:       model.Title,
		Address:     model.Address,
		Description: model.Description,
		Drinks:      drinks,
		CreatedAt:   timestamppb.New(model.CreatedAt.Time),
		UpdatedAt:   timestamppb.New(model.UpdatedAt.Time),
	}, nil
}

func (s *BarMapService) CreateBar(ctx context.Context, req *CreateBarRequest) (*Bar, error) {

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

func (s *BarMapService) UpdateBar(ctx context.Context, req *UpdateBarRequest) (*Bar, error) {
	var now = time.Now().Format("2006-01-02 15:04:05.000000")

	tx, err := s.handler.conn.BeginEx(ctx, nil)
	if err != nil {
		log.Println("Failed start bar update sql transaction: ", err)
		return &Bar{}, err
	}

	// var sql string = "begin transation;"

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

	// sql += fmt.Sprintf("update bars set title = '%s', address = '%s', description = '%s', drinks_id = '%s', "+
	// 	"updated_at = '%s' where id = %s returning *;",
	// 	req.Title, req.Address, req.Description, GetDrinksIds(req.Drinks), now, req.Id)

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
		Id:          fmt.Sprintf("%d", model.Id),
		Title:       model.Title,
		Address:     model.Address,
		Description: model.Description,
		Drinks:      drinks,
		CreatedAt:   timestamppb.New(model.CreatedAt.Time),
		UpdatedAt:   timestamppb.New(model.UpdatedAt.Time),
	}, nil
}

func (s *BarMapService) DeleteBar(ctx context.Context, req *DeleteBarRequest) (*DeleteBarResponse, error) {
	sql := "begin transaction;"
	sql += fmt.Sprintf("delete from bars WHERE id = %s;", req.Id)
	sql += "commit;"
	_, err := s.handler.conn.Exec(sql)
	if err != nil {
		log.Println("Bar deleting transaction failed")
		_, err := s.handler.conn.Exec("rollback;")
		if err != nil {
			log.Println("Rollback bar deleting transactionfailed")
			panic(err)
		}
		return &DeleteBarResponse{}, err
	}
	return &DeleteBarResponse{}, nil
}

func (s *BarMapService) ListBar(ctx context.Context, req *ListBarsRequest) (*ListBarsResponse, error) {
	return &ListBarsResponse{}, nil
}

func (s *BarMapService) GetBar(ctx context.Context, req *GetBarRequest) (*Bar, error) {

	var sql string = fmt.Sprintf("select * from bars where id = %s;", req.Id)
	return WrapBarResponse(s.handler.conn.Query(sql))
}

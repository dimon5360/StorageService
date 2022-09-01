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

/// postgres response wrapper
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

/// gRPC creating bar request handler
func (s *BarMapService) CreateBar(ctx context.Context, req *CreateBarRequest) (*Bar, error) {

	var now = time.Now().Format("2006-01-02 15:04:05.000000")

	var sql string = fmt.Sprintf("insert into bars "+
		"(title, address, description, drinks_id, created_at, updated_at) "+
		"values ('%s', '%s', '%s', '%s', '%s', '%s') returning *;",
		req.Title, req.Address, req.Description, "{}", now, now) // drinks does not fill now

	return WrapBarResponse(s.handler.conn.Query(sql))
}

/// gRPC updating bar request handler
func (s *BarMapService) UpdateBar(ctx context.Context, req *UpdateBarRequest) (*Bar, error) {

	var now = time.Now().Format("2006-01-02 15:04:05.000000")

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
		"updated_at = '%s' where id = %s;",
		req.Title, req.Address, req.Description, GetDrinksIds(req.Drinks), now, req.Id)

	for _, drink := range req.Drinks {

		sql += fmt.Sprintf("update drinks set title = '%s', price = '%s', type = '%s', description = '%s', "+
			"bar_id = '%s', updated_at = '%s' where id = %s and bar_id = %s;",
			drink.Title, drink.Price, drink.DrinkType, req.Description, drink.BarId, now, drink.Id, req.Id)

		for _, ingredient := range drink.Ingredients {
			sql += fmt.Sprintf("update ingredients set title = '%s', amount = '%s', "+
				"drink_id = '%s', updated_at = '%s' where id = %s and drink_id = %s;",
				ingredient.Title, ingredient.Amount, ingredient.DrinkId, now, ingredient.Id, drink.Id)
		}
	}

	sql += "commit;"
	sql += fmt.Sprintf("select * from bars where id = %s;", req.Id)

	return WrapBarResponse(s.handler.conn.Query(sql))
}

/// gRPC deleting bar request handler
func (s *BarMapService) DeleteBar(ctx context.Context, req *DeleteBarRequest) (*DeleteBarResponse, error) {
	sql := fmt.Sprintf("delete from bars WHERE id = %s;", req.Id)
	_, err := s.handler.conn.Exec(sql)
	if err != nil {
		log.Println("Bar deleting transaction failed")
		return &DeleteBarResponse{}, err
	}
	return &DeleteBarResponse{}, nil
}

/// gRPC getting bars list request handler
func (s *BarMapService) ListBar(ctx context.Context, req *ListBarsRequest) (*ListBarsResponse, error) {
	return &ListBarsResponse{}, nil
}

/// gRPC getting bar request handler
func (s *BarMapService) GetBar(ctx context.Context, req *GetBarRequest) (*Bar, error) {

	var sql string = fmt.Sprintf("select * from bars where id = %s;", req.Id)
	return WrapBarResponse(s.handler.conn.Query(sql))
}

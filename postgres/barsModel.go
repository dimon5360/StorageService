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
	for _, idx := range model.Drinks_Id.Elements {
		drinks = append(drinks, &Drink{
			Id: fmt.Sprintf("%d", idx.Int),
		})
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

/// gRPC creating bar request handler
/// established, remove comment later
func (s *BarMapService) CreateBar(ctx context.Context, req *CreateBarRequest) (*Bar, error) {

	var now = time.Now().Format("2006-01-02 15:04:05.000000")

	var sql string = fmt.Sprintf("insert into bars "+
		"(title, address, description, created_at, updated_at) "+
		"values ('%s', '%s', '%s', '%s', '%s') returning *;",
		req.Title, req.Address, req.Description, now, now) // drinks does not fill now

	return WrapBarResponse(s.handler.conn.Query(sql))
}

/// gRPC updating bar request handler
/// established, remove comment later
func (s *BarMapService) UpdateBar(ctx context.Context, req *UpdateBarRequest) (*Bar, error) {

	var now = time.Now().Format("2006-01-02 15:04:05.000000")

	var sql string = fmt.Sprintf("update bars set title = '%s', address = '%s', description = '%s', "+
		"updated_at = '%s' where id = %s returning *",
		req.Title, req.Address, req.Description, now, req.Id)

	return WrapBarResponse(s.handler.conn.Query(sql))
}

/// gRPC deleting bar request handler
/// established, remove comment later
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
// #TODO: need to implement script
func (s *BarMapService) ListBar(ctx context.Context, req *ListBarsRequest) (*ListBarsResponse, error) {
	return &ListBarsResponse{}, nil
}

/// gRPC getting bar request handler
/// established, remove comment later
// return only drinks id, drinks can be asked separately
func (s *BarMapService) GetBar(ctx context.Context, req *GetBarRequest) (*Bar, error) {
	var sql string = fmt.Sprintf("select * from bars where id = %s;", req.Id)
	return WrapBarResponse(s.handler.conn.Query(sql))
}

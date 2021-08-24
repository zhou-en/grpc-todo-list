package v1

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes"
	_ "github.com/lib/pq"
	"github.com/zhou-en/grpc-todo-list/pkg/api/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// apiVersion is version of API is provided by server
	apiVersion = "v1"
)

// toDoServiceServer is implementation of v1.ToDoServiceServer proto interface
type toDoServiceServer struct {
	db *sql.DB
}

// Create new todo task
func (s *toDoServiceServer) Create(ctx context.Context, req *v1.CreateRequest) (*v1.CreateResponse, error) {
	// check if the API version requested by client is supported by server
	if err := s.checkAPI(req.Api); err != nil {
		return nil, err
	}
	// get SQL connection from pool
	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	reminder, err := ptypes.Timestamp(req.ToDo.Reminder)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "reminder field has invlid format -> " + err.Error())
	}

	// Insert ToDo entity data
	lastInsertId := 0
	//res, err := c.ExecContext(ctx, `INSERT INTO "ToDo" ("Title", "Description", "Reminder") VALUES($1, $2, $3)`,
	//	req.ToDo.Title, req.ToDo.Description, reminder)
	//
	err = s.db.QueryRow(`INSERT INTO ToDo (Title, Description, Reminder) VALUES($1, $2, $3) RETURNING ID`, req.ToDo.Title, req.ToDo.Description, reminder).Scan(&lastInsertId)


	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to insert into ToDo -> " + err.Error())
	}

	return &v1.CreateResponse{
		Api: apiVersion,
		Id: int64(lastInsertId),
	}, nil
}

func (s *toDoServiceServer) Read(ctx context.Context, req *v1.ReadRequest) (*v1.ReadResponse, error) {
	// check API version
	if err := s.checkAPI(req.Api); err != nil {
		return nil, err
	}
	// get sql  connection
	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}
	// query todo
	rows, err := c.QueryContext(ctx, `SELECT ID, Title, Description, Reminder FROM ToDo WHERE ID=$1`, req.Id)
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to select from ToDo table -> " + err.Error())
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, status.Error(codes.Unknown, "failed to retrieve data from ToDo -> " + err.Error())
		}
		return nil, status.Error(codes.NotFound, fmt.Sprintf("ToDo with ID = '%d' is not found", req.Id))
	}

	// get ToDo data
	var td v1.ToDo
	var reminder time.Time
	if err := rows.Scan(&td.Id, &td.Title, &td.Description, &reminder); err != nil {
		return nil, status.Error(codes.Unknown, "failed to retrive field values from ToDo row -> " + err.Error())
	}
	td.Reminder, err = ptypes.TimestampProto(reminder)
	if err != nil {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("found multiple ToDo with ID = '%d'", req.Id))
	}

	return &v1.ReadResponse{
		Api: apiVersion,
		ToDo: &td,
	}, nil

}

func (s *toDoServiceServer) Update(ctx context.Context, req *v1.UpdateRequest) (*v1.UpdateResponse, error) {
	// check api
	if err := s.checkAPI(req.Api); err != nil {
		return nil, err
	}

	// get sql connection
	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	reminder, err := ptypes.Timestamp(req.ToDo.Reminder)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "reminder field has invalid format -> " + err.Error())
	}

	// update todo
	sqlQuery := `UPDATE ToDo SET Title=$1, Description=$2, Reminder=$3 WHERE ID=$4;`
	res, err := c.ExecContext(ctx, sqlQuery, req.ToDo.Title, req.ToDo.Description, reminder, req.ToDo.Id)
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to update ToDo -> " + err.Error())
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to retrieve rows affected value -> " + err.Error())
	}
	if rows == 0 {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("ToDo with ID='%d' is not found", req.ToDo.Id))
	}

	return &v1.UpdateResponse{
		Api: apiVersion,
		Updated: rows,
	}, nil
}

func (s *toDoServiceServer) Delete(ctx context.Context, req *v1.DeleteRequest) (*v1.DeleteResponse, error) {
	if err := s.checkAPI(req.Api); err != nil {
		return nil, err
	}

	// get sql connection
	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	// delete ToDo
	deleteQuery := `DELETE FROM ToDo WHERE ID=$1`
	res, err := c.ExecContext(ctx, deleteQuery, req.Id)
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to delete ToDo -> " + err.Error())
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to retrieve rows affected")
	}
	if rows == 0 {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("ToDo with ID='%d' is not found", req.Id))
	}

	return &v1.DeleteResponse{
		Api: apiVersion,
		Deleted: rows,
	}, nil
}

func (s *toDoServiceServer) ReadAll(ctx context.Context, req *v1.ReadAllRequest) (*v1.ReadAllResponse, error) {
	if err := s.checkAPI(req.Api); err != nil {
		return nil, err
	}

	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	allQuery := `SELECT ID, Title, Description, Reminder FROM ToDo;`
	rows, err := c.QueryContext(ctx, allQuery)
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to select from ToDo -> " + err.Error())
	}
	defer rows.Close()

	var reminder time.Time
	list := []*v1.ToDo{}
	for rows.Next() {
		td := new(v1.ToDo)
		if err := rows.Scan(&td.Id, &td.Title, &td.Description, &reminder); err != nil {
			return nil, status.Error(codes.Unknown, "failed to retrieve field values from ToDo row -> " + err.Error())
		}
		td.Reminder, err = ptypes.TimestampProto(reminder)
		if err != nil {
			return nil, status.Error(codes.Unknown, "reminder field has invalid format -> " + err.Error())
		}
	}
	return &v1.ReadAllResponse{
		Api: apiVersion,
		ToDos: list,
	}, nil
}

// NewToDoServiceServer creates ToDo service
func NewToDoServiceServer(db *sql.DB) v1.ToDoServiceServer {
	return &toDoServiceServer{db: db}
}

// checkAPI checks if the API version requested by client is supported by server
func (s *toDoServiceServer) checkAPI(api string) error {
	// API version is "" means use current version of the service
	if len(api) > 0 {
		if apiVersion != api {
			return status.Errorf(codes.Unimplemented, "unsupported API version: service implements API version '%s', but asked for '%s'", apiVersion, api)
		}
	}
	return nil
}

// connect returns SQL database connection from the pool
func (s *toDoServiceServer) connect(ctx context.Context) (*sql.Conn, error) {
	c, err := s.db.Conn(ctx)
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to connect to database -> " + err.Error())
	}
	return c, nil
}


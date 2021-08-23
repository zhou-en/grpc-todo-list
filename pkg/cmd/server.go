package cmd

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/zhou-en/grpc-todo-list/pkg/protocol/grpc"
	"github.com/zhou-en/grpc-todo-list/pkg/service/v1"
	"log"
)

//  Config for server
type Config struct {
	// gRPC server start parameters section
	// gRPC is a TCP port to listen by gRPC server
	GRPCPort string
	// DB parameters
	DatastoreDBHost string
	DatastoreDBUser string
	DatastoreDBPassword string
	DatastoreDBSchema string
}

// RunServer runs gRPC server and HTTP gateway
func RunServer() error {
	ctx := context.Background()

	// get config
	var cfg Config
	flag.StringVar(&cfg.GRPCPort, "grpc-port", "9090", "gRPC port to bind")
	flag.StringVar(&cfg.DatastoreDBHost, "db-host", "localhost", "Database host")
	flag.StringVar(&cfg.DatastoreDBUser, "db-user", "postgres", "Database user")
	flag.StringVar(&cfg.DatastoreDBPassword, "db-password", "postgres", "Database password")
	flag.StringVar(&cfg.DatastoreDBSchema, "db-schema", "todo_list", "Database schema")
	flag.Parse()

	if len(cfg.GRPCPort) == 0 {
		return fmt.Errorf("Invalid TCP port for gRPC server: '%s'", cfg.GRPCPort)
	}

	// Database
	//param := "parseTime=true"
	//dsn := fmt.Sprintf("postgres://%s:%s@%s:5432/%s?%s",
	//	cfg.DatastoreDBUser,
	//	cfg.DatastoreDBPassword,
	//	cfg.DatastoreDBHost,
	//	cfg.DatastoreDBSchema,
	//	param)
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		cfg.DatastoreDBHost, 5432, cfg.DatastoreDBUser, cfg.DatastoreDBPassword, cfg.DatastoreDBSchema)

	log.Println(psqlInfo)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return fmt.Errorf("Failed to open database: %v", err)
	}
	defer db.Close()

	v1Api := v1.NewToDoServiceServer(db)
	return grpc.RunServer(ctx, v1Api, cfg.GRPCPort)

}


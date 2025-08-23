package db

import (
	"context"
	"fmt"
	"github.com/dresswithpockets/openstats/app/db/query"
	"github.com/dresswithpockets/openstats/app/env"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	pgxUUID "github.com/vgarvardt/pgx-google-uuid/v5"
	"log"
)

var Queries *query.Queries
var DB *Actions

func SetupDB(ctx context.Context) error {
	var (
		dbHost     = env.GetString("OPENSTATS_DB_HOST")
		dbPort     = env.GetString("OPENSTATS_DB_PORT")
		dbName     = env.GetString("OPENSTATS_DB_DATABASE")
		dbUsername = env.GetString("OPENSTATS_DB_USERNAME")
		dbPassword = env.GetString("OPENSTATS_DB_PASSWORD")
		dbTraceLog = env.GetBool("OPENSTATS_DB_TRACE_LOG")
	)

	connectionString := fmt.Sprintf(
		"host=%s port=%s database=%s user=%s password=%s",
		dbHost,
		dbPort,
		dbName,
		dbUsername,
		dbPassword,
	)

	config, configErr := pgxpool.ParseConfig(connectionString)
	if configErr != nil {
		return configErr
	}

	if dbTraceLog {
		config.ConnConfig.Tracer = &tracelog.TraceLog{
			Logger: tracelog.LoggerFunc(func(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]interface{}) {
				log.Printf("[%s] %s %v", level, msg, data)
			}),
			LogLevel: tracelog.LogLevelTrace,
		}
	}

	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		pgxUUID.Register(conn.TypeMap())
		return nil
	}

	pool, poolErr := pgxpool.NewWithConfig(ctx, config)
	if poolErr != nil {
		return poolErr
	}

	Queries = query.New(pool)
	DB = NewActions(pool, Queries)
	return nil
}

package db

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dresswithpockets/openstats/app/db/query"
	pgxUUID "github.com/vgarvardt/pgx-google-uuid/v5"
)

var Queries *query.Queries
var DB *Actions

func SetupDB(ctx context.Context) error {
	config, configErr := pgxpool.ParseConfig("host=localhost port=15432 database=openstats user=openstats password=openstats")
	if configErr != nil {
		return configErr
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

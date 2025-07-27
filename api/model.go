package main

import (
	"context"
	"github.com/dresswithpockets/openstats/app/db"
	"github.com/dresswithpockets/openstats/app/db/query"
	"github.com/jackc/pgx/v5"
)

var Queries *query.Queries
var Actions *db.Actions

func SetupDB(ctx context.Context) error {
	conn, connErr := pgx.Connect(ctx, "host=localhost port=15432 database=openstats user=openstats password=openstats")
	if connErr != nil {
		return connErr
	}

	Queries = query.New(conn)
	Actions = db.NewActions(conn, Queries)
	return nil
}

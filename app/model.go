package main

import (
	"context"
	"github.com/jackc/pgx/v5"

	"github.com/dresswithpockets/openstats/app/queries"
	"github.com/dresswithpockets/openstats/app/query"
)

var Queries *query.Queries
var Actions *queries.Actions

func SetupDB(ctx context.Context) error {
	conn, connErr := pgx.Connect(ctx, "host=localhost port=15432 database=openstats user=openstats password=openstats")
	if connErr != nil {
		return connErr
	}

	Queries = query.New(conn)
	Actions = queries.NewActions(conn, Queries)
	return nil
}

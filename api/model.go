package main

import (
	"context"

	"github.com/dresswithpockets/openstats/app/db"
	"github.com/dresswithpockets/openstats/app/db/query"
	"github.com/jackc/pgx/v5"
	pgxUUID "github.com/vgarvardt/pgx-google-uuid/v5"
)

var Queries *query.Queries
var Actions *db.Actions

func SetupDB(ctx context.Context) error {
	// TODO: switch to pgxpool
	conn, connErr := pgx.Connect(ctx, "host=localhost port=15432 database=openstats user=openstats password=openstats")
	if connErr != nil {
		return connErr
	}

	pgxUUID.Register(conn.TypeMap())

	Queries = query.New(conn)
	Actions = db.NewActions(conn, Queries)
	return nil
}

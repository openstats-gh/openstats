package main

import (
	"database/sql"

	"github.com/dresswithpockets/openstats/app/queries"
	"github.com/dresswithpockets/openstats/app/query"
)

var Queries *query.Queries
var Actions *queries.Actions

func SetupDB() error {
	db, dbErr := sql.Open("sqlite3", "local.openstats.db")
	if dbErr != nil {
		return dbErr
	}

	Queries = query.New(db)
	Actions = queries.NewActions(db, Queries)
	return nil
}

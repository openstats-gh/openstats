package queries

import (
	"context"
	"database/sql"
	"errors"
	"github.com/dresswithpockets/openstats/app/query"
	"github.com/mattn/go-sqlite3"
)

var ErrSlugAlreadyInUse = errors.New("slug already in use")

type Actions struct {
	queries *query.Queries
	db      *sql.DB
}

func NewActions(db *sql.DB, queries *query.Queries) *Actions {
	return &Actions{
		queries: queries,
		db:      db,
	}
}

func (a *Actions) CreateUser(ctx context.Context, slug, encodedPasswordHash, email, displayName string) (*query.User, error) {
	tx, txErr := a.db.BeginTx(ctx, nil)
	if txErr != nil {
		return nil, txErr
	}

	//goland:noinspection GoUnhandledErrorResult
	defer tx.Rollback()

	qtx := a.queries.WithTx(tx)
	user, createUserErr := qtx.CreateUser(ctx, slug)
	if errors.Is(createUserErr, sqlite3.ErrConstraintUnique) {
		return nil, ErrSlugAlreadyInUse
	}

	if createUserErr != nil {
		return nil, createUserErr
	}

	if err := qtx.AddUserSlugRecord(ctx, query.AddUserSlugRecordParams{
		UserID: user.ID,
		Slug:   slug,
	}); err != nil {
		return nil, err
	}

	if err := qtx.CreateUserPassword(ctx, query.CreateUserPasswordParams{
		UserID:      user.ID,
		EncodedHash: encodedPasswordHash,
	}); err != nil {
		return nil, err
	}

	if len(email) > 0 {
		if err := qtx.AddUserEmail(ctx, query.AddUserEmailParams{
			UserID: user.ID,
			Email:  email,
		}); err != nil {
			return nil, err
		}
	}

	if len(displayName) > 0 {
		if err := qtx.AddUserDisplayName(ctx, query.AddUserDisplayNameParams{
			UserID:      user.ID,
			DisplayName: displayName,
		}); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &user, nil
}

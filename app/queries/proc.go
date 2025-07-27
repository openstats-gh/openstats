package queries

import (
	"context"
	"errors"
	"github.com/dresswithpockets/openstats/app/query"
	"github.com/jackc/pgx/v5"
	"github.com/mattn/go-sqlite3"
)

var ErrSlugAlreadyInUse = errors.New("slug already in use")

type Actions struct {
	queries *query.Queries
	conn    *pgx.Conn
}

func NewActions(conn *pgx.Conn, queries *query.Queries) *Actions {
	return &Actions{
		queries: queries,
		conn:    conn,
	}
}

func (a *Actions) CreateUser(ctx context.Context, slug, encodedPasswordHash, email, displayName string) (*query.User, error) {
	tx, txErr := a.conn.BeginTx(ctx, pgx.TxOptions{})
	if txErr != nil {
		return nil, txErr
	}

	//goland:noinspection GoUnhandledErrorResult
	defer tx.Rollback(ctx)

	qtx := a.queries.WithTx(tx)
	user, createUserErr := qtx.AddUser(ctx, slug)
	if errors.Is(createUserErr, sqlite3.ErrConstraintUnique) {
		return nil, ErrSlugAlreadyInUse
	}

	if createUserErr != nil {
		return nil, createUserErr
	}

	if err := qtx.AddUserSlugHistory(ctx, query.AddUserSlugHistoryParams{
		UserID: user.ID,
		Slug:   slug,
	}); err != nil {
		return nil, err
	}

	if err := qtx.AddUserPassword(ctx, query.AddUserPasswordParams{
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

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &user, nil
}

package db

import (
	"context"
	"errors"
	"github.com/Masterminds/squirrel"
	"github.com/dresswithpockets/openstats/app/db/query"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrSlugAlreadyInUse = errors.New("slug already in use")

type Actions struct {
	pool    *pgxpool.Pool
	queries *query.Queries
}

func NewActions(pool *pgxpool.Pool, queries *query.Queries) *Actions {
	return &Actions{
		queries: queries,
		pool:    pool,
	}
}

func (a *Actions) Builder() squirrel.StatementBuilderType {
	return squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
}

func (a *Actions) Exec(ctx context.Context, sqlizer squirrel.Sqlizer) (pgconn.CommandTag, error) {
	sql, args, sqlErr := sqlizer.ToSql()
	if sqlErr != nil {
		return pgconn.CommandTag{}, sqlErr
	}

	return a.pool.Exec(ctx, sql, args...)
}

func (a *Actions) Query(ctx context.Context, sqlizer squirrel.Sqlizer) (pgx.Rows, error) {
	sql, args, sqlErr := sqlizer.ToSql()
	if sqlErr != nil {
		return nil, sqlErr
	}

	return a.pool.Query(ctx, sql, args...)
}

func (a *Actions) QueryRow(ctx context.Context, sqlizer squirrel.Sqlizer) (pgx.Row, error) {
	sql, args, sqlErr := sqlizer.ToSql()
	if sqlErr != nil {
		return nil, sqlErr
	}

	return a.pool.QueryRow(ctx, sql, args...), nil
}

func (a *Actions) ScanRow(ctx context.Context, sqlizer squirrel.Sqlizer, dest ...any) error {
	sql, args, sqlErr := sqlizer.ToSql()
	if sqlErr != nil {
		return sqlErr
	}

	return a.pool.QueryRow(ctx, sql, args...).Scan(dest...)
}

func (a *Actions) CreateUser(ctx context.Context, slug, encodedPasswordHash, email, displayName string) (*query.User, error) {
	tx, txErr := a.pool.BeginTx(ctx, pgx.TxOptions{})
	if txErr != nil {
		return nil, txErr
	}

	//goland:noinspection GoUnhandledErrorResult
	defer tx.Rollback(ctx)

	qtx := a.queries.WithTx(tx)
	user, createUserErr := qtx.AddUser(ctx, slug)

	if IsUniqueConstraintErr(createUserErr) {
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

//type CreateUserParams struct {
//	Slug                string
//	EncodedPasswordHash string
//	Email               string
//	DisplayName         string
//}
//
//func (a *Actions) CreateUsers(ctx context.Context, users []CreateUserParams) ([]uuid.UUID, error) {
//	tx, txErr := a.pool.BeginTx(ctx, pgx.TxOptions{})
//	if txErr != nil {
//		return nil, txErr
//	}
//
//	//goland:noinspection GoUnhandledErrorResult
//	defer tx.Rollback(ctx)
//
//	var slugs []string
//	for _, user := range users {
//		slugs = append(slugs, user.Slug)
//	}
//	if _, err := Queries.AddUsers(ctx, slugs); err != nil {
//		return nil, err
//	}
//
//	uuids, uuidsErr := Queries.FindUserUUIDsBySlugs(ctx, slugs)
//	if uuidsErr != nil {
//		return nil, uuidsErr
//	}
//
//	Queries.AddUserSlugHistories(ctx)
//
//	if err := tx.Commit(ctx); err != nil {
//		return nil, err
//	}
//
//	return uuids, nil
//}

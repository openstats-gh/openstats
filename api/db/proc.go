package db

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/dresswithpockets/openstats/app/db/query"
	"github.com/dresswithpockets/openstats/app/rid"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rotisserie/eris"
	"time"
)

var ErrSlugAlreadyInUse = eris.New("slug already in use")

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

type CreatedUser struct {
	User  query.User
	Email query.UserEmail
}

func (a *Actions) CreateUser(ctx context.Context, slug, encodedPasswordHash, email, displayName string) (*CreatedUser, error) {
	tx, err := a.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, eris.Wrap(err, "error beginning transaction")
	}

	//goland:noinspection GoUnhandledErrorResult
	defer tx.Rollback(ctx)

	qtx := a.queries.WithTx(tx)

	var createdUser CreatedUser
	createdUser.User, err = qtx.AddUser(ctx, slug)
	if IsUniqueConstraintErr(err) {
		return nil, ErrSlugAlreadyInUse
	}

	if err != nil {
		return nil, eris.Wrap(err, "error adding user to db")
	}

	if err = qtx.AddUserSlugHistory(ctx, query.AddUserSlugHistoryParams{
		UserID: createdUser.User.ID,
		Slug:   slug,
	}); err != nil {
		return nil, eris.Wrap(err, "error adding slug to history")
	}

	if err = qtx.AddUserPassword(ctx, query.AddUserPasswordParams{
		UserID:      createdUser.User.ID,
		EncodedHash: encodedPasswordHash,
	}); err != nil {
		return nil, eris.Wrap(err, "error adding user password")
	}

	if len(email) > 0 {
		createdUser.Email, err = qtx.AddOrGetUserEmail(ctx, query.AddOrGetUserEmailParams{
			UserID:    createdUser.User.ID,
			Email:     email,
			OtpSecret: rand.Text(),
		})
		if err != nil {
			return nil, eris.Wrap(err, "error adding user email")
		}
	}

	if len(displayName) > 0 {
		if err = qtx.AddUserDisplayName(ctx, query.AddUserDisplayNameParams{
			UserID:      createdUser.User.ID,
			DisplayName: displayName,
		}); err != nil {
			return nil, eris.Wrap(err, "error adding user display name")
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, eris.Wrap(err, "error committing transaction")
	}

	return &createdUser, nil
}

func (a *Actions) CreateGameSessionAndToken(
	ctx context.Context,
	gameToken uuid.UUID,
	userRid, gameRid rid.RID,
	issuer, audience string,
	duration time.Duration,
	jitter time.Duration,
) (token query.Token, session query.GameSession, err error) {
	var tx pgx.Tx
	tx, err = a.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return
	}

	//goland:noinspection GoUnhandledErrorResult
	defer tx.Rollback(ctx)

	qtx := a.queries.WithTx(tx)
	session, err = qtx.CreateGameSession(ctx, query.CreateGameSessionParams{
		GameUuid:      gameRid.ID,
		UserUuid:      userRid.ID,
		GameTokenUuid: gameToken,
	})
	if err != nil {
		return
	}

	sessionRid := rid.From("gs", session.Uuid)
	subject := fmt.Sprintf("users/v1/%s/games/%s/sessions/%s", userRid.String(), gameRid.String(), sessionRid.String())

	nowTime := time.Now().UTC()
	token, err = qtx.CreateToken(ctx, query.CreateTokenParams{
		Issuer:    issuer,
		Subject:   subject,
		Audience:  audience,
		ExpiresAt: nowTime.Add(duration),
		NotBefore: nowTime.Add(-jitter),
		IssuedAt:  nowTime,
	})
	if err != nil {
		return
	}

	if err = tx.Commit(ctx); err != nil {
		return
	}

	return
}

func (a *Actions) Transact(ctx context.Context, do func(context.Context, *query.Queries) error) (err error) {
	var tx pgx.Tx
	tx, err = a.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return
	}

	qtx := a.queries.WithTx(tx)
	if err = do(ctx, qtx); err != nil {
		return tx.Rollback(ctx)
	}

	if err = tx.Commit(ctx); err != nil {
		return tx.Rollback(ctx)
	}

	return nil
}

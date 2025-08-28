package internal

import (
	"context"
	"database/sql"
	"errors"
	"github.com/dresswithpockets/openstats/app/db"
	"github.com/dresswithpockets/openstats/app/db/query"
	"github.com/rotisserie/eris"
	"strconv"
)

type SendSlugReminderInput struct {
	Body struct {
		Email string `json:"email" format:"email"`
	}
}
type SendSlugReminderOutput struct{}

func HandleSendSlugReminder(ctx context.Context, input *SendSlugReminderInput) (*SendSlugReminderOutput, error) {
	slugs, dbErr := db.Queries.GetSlugsByEmail(ctx, input.Body.Email)
	if errors.Is(dbErr, sql.ErrNoRows) {
		return &SendSlugReminderOutput{}, nil
	}

	if dbErr != nil {
		return nil, dbErr
	}

	if len(slugs) == 0 {
		return &SendSlugReminderOutput{}, nil
	}

	err := SendSlugReminder(ctx, input.Body.Email, slugs)
	if err != nil {
		return nil, err
	}

	return &SendSlugReminderOutput{}, nil
}

type SendPasswordResetInput struct {
	Body struct {
		Slug string `json:"slug"`
	}
}

type SendPasswordResetOutput struct{}

func HandleSendPasswordReset(ctx context.Context, input *SendPasswordResetInput) (*SendPasswordResetOutput, error) {
	userEmail, err := db.Queries.FindUserEmailBySlug(ctx, input.Body.Slug)
	if eris.Is(err, sql.ErrNoRows) {
		return &SendPasswordResetOutput{}, nil
	}

	if err != nil {
		return nil, eris.Wrap(err, "error sending password reset email")
	}

	var hmacSecret string
	hmacSecret, err = db.Queries.SecretRead(ctx, query.SecretReadParams{
		Path: db.PrivateUser2faHmacSecretPath,
		Key:  strconv.FormatInt(int64(userEmail.UserID), 10),
	})

	if err != nil {
		return nil, eris.Wrap(err, "there was an error creating your 2FA TOTP code")
	}

	err = Send2faTotpEmail(ctx, PasswordResetPurpose, input.Body.Slug, hmacSecret, userEmail.Email)
	if err != nil {
		return nil, eris.Wrap(err, "error sending password reset email")
	}

	return &SendPasswordResetOutput{}, nil
}

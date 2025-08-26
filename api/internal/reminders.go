package internal

import (
	"context"
	"database/sql"
	"errors"
	"github.com/dresswithpockets/openstats/app/db"
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

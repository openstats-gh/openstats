package internal

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/dresswithpockets/openstats/app/db"
	"github.com/dresswithpockets/openstats/app/db/query"
	"github.com/dresswithpockets/openstats/app/env"
	"github.com/dresswithpockets/openstats/app/mail"
	"github.com/google/uuid"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"net/url"
	"time"
)

var ValidateOptions = totp.ValidateOpts{
	Period:    30 * 60,
	Digits:    6,
	Algorithm: otp.AlgorithmSHA512,
}

func AddUserEmail(ctx context.Context, userUuid uuid.UUID, email string) error {
	dbUserEmail, dbErr := db.Queries.AddOrGetUserEmail(ctx, query.AddOrGetUserEmailParams{
		UserUuid:  userUuid,
		Email:     email,
		OtpSecret: rand.Text(),
	})
	if dbErr != nil {
		return dbErr
	}

	totpCode, totpErr := totp.GenerateCodeCustom(dbUserEmail.OtpSecret, time.Now(), ValidateOptions)

	if totpErr != nil {
		return totpErr
	}

	appBaseUrl := env.GetString("OPENSTATS_APP_BASEURL")
	confUrl := fmt.Sprintf("%s/confirm-email?e=%s&c=%s", appBaseUrl, url.QueryEscape(email), url.QueryEscape(totpCode))
	confBody := fmt.Sprintf("Confirm adding your email address to your Openstats account by clicking on the link below.<br/></br><a href=\"%s\">%s</a>", confUrl, confUrl)

	return mail.Default.Send(ctx, mail.Mail{
		From:    "noreply@openstats.me",
		To:      email,
		Subject: "Openstats Confirmation",
		Body:    confBody,
	})
}

func ValidateUserEmail(ctx context.Context, userUuid uuid.UUID, email, passcode string) (bool, error) {
	dbUserEmail, dbErr := db.Queries.GetUserEmail(ctx, query.GetUserEmailParams{
		UserUuid: userUuid,
		Email:    email,
	})

	if dbErr != nil {
		return false, dbErr
	}

	return totp.ValidateCustom(passcode, dbUserEmail.OtpSecret, time.Now(), ValidateOptions)
}

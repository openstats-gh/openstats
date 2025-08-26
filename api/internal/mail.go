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
	"github.com/rotisserie/eris"
	"net/url"
	"time"
)

var ValidateOptions = totp.ValidateOpts{
	Period:    30 * 60,
	Digits:    6,
	Algorithm: otp.AlgorithmSHA512,
}

func SendEmailConfirmation(ctx context.Context, userEmail query.UserEmail) error {
	totpCode, totpErr := totp.GenerateCodeCustom(userEmail.OtpSecret, time.Now(), ValidateOptions)

	if totpErr != nil {
		return totpErr
	}

	appBaseUrl := env.GetString("OPENSTATS_APP_BASEURL")
	confUrl := fmt.Sprintf("%s/confirm-email?e=%s&c=%s", appBaseUrl, url.QueryEscape(userEmail.Email), url.QueryEscape(totpCode))
	confBody := fmt.Sprintf("Confirm adding your email address to your Openstats account by clicking on the link below.<br/></br><a href=\"%s\">%s</a>", confUrl, confUrl)

	return mail.Default.Send(ctx, mail.Mail{
		From:    "noreply@openstats.me",
		To:      userEmail.Email,
		Subject: "Openstats Confirmation",
		Body:    confBody,
	})
}

func SendSlugReminder(ctx context.Context, email string, slugs []string) error {
	var listItems string
	for _, slug := range slugs {
		listItems += "<li>" + slug
	}

	listItems += "</li>"

	body := fmt.Sprintf("Below are the slugs associated with your email. Each of these slugs can be used to sign into a different account:<br><br><ul>%s</ul>", listItems)
	return mail.Default.Send(ctx, mail.Mail{
		From:    "noreply@openstats.me",
		To:      email,
		Subject: "Your Openstats Slugs",
		Body:    body,
	})
}

func AddUserEmailAndSendConfirmation(ctx context.Context, userUuid uuid.UUID, email string) error {
	userEmail, err := db.Queries.AddOrGetUserEmailByUuid(ctx, query.AddOrGetUserEmailByUuidParams{
		UserUuid:  userUuid,
		Email:     email,
		OtpSecret: rand.Text(),
	})
	if err != nil {
		return eris.Wrap(err, "error adding user email")
	}

	totpCode, totpErr := totp.GenerateCodeCustom(userEmail.OtpSecret, time.Now(), ValidateOptions)

	if totpErr != nil {
		return totpErr
	}

	appBaseUrl := env.GetString("OPENSTATS_APP_BASEURL")
	confUrl := fmt.Sprintf("%s/confirm-email?e=%s&c=%s", appBaseUrl, url.QueryEscape(userEmail.Email), url.QueryEscape(totpCode))
	confBody := fmt.Sprintf("Confirm adding your email address to your Openstats account by clicking on the link below.<br/></br><a href=\"%s\">%s</a>", confUrl, confUrl)

	return mail.Default.Send(ctx, mail.Mail{
		From:    "noreply@openstats.me",
		To:      userEmail.Email,
		Subject: "Openstats Confirmation",
		Body:    confBody,
	})
}

func ValidateUserEmail(ctx context.Context, userId int32, email, passcode string) (bool, error) {
	dbUserEmail, dbErr := db.Queries.GetUserEmail(ctx, query.GetUserEmailParams{
		UserID: userId,
		Email:  email,
	})

	if dbErr != nil {
		return false, eris.Wrap(dbErr, "error getting user email")
	}

	validated, validateErr := totp.ValidateCustom(passcode, dbUserEmail.OtpSecret, time.Now(), ValidateOptions)
	if validateErr != nil {
		return false, eris.Wrap(validateErr, "error validating OTP")
	}

	_, dbErr = db.Queries.ConfirmEmail(ctx, query.ConfirmEmailParams{
		UserID: userId,
		Email:  "",
	})

	if dbErr != nil {
		return false, eris.Wrap(dbErr, "confirming user email in db")
	}

	return validated, nil
}

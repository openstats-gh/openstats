package internal

import (
	"context"
	"fmt"
	"github.com/dresswithpockets/openstats/app/auth"
	"github.com/dresswithpockets/openstats/app/db"
	"github.com/dresswithpockets/openstats/app/db/query"
	"github.com/dresswithpockets/openstats/app/env"
	"github.com/dresswithpockets/openstats/app/mail"
	"github.com/pquerna/otp/totp"
	"github.com/rotisserie/eris"
	"net/url"
	"strconv"
	"time"
)

func SendEmailConfirmation(ctx context.Context, otpSecret, email string) error {
	totpCode, totpErr := totp.GenerateCodeCustom(otpSecret, time.Now(), auth.ValidateOptions)

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

type TotpPurpose int

const (
	PasswordResetPurpose TotpPurpose = iota
	EmailConfirmationPurpose
)

func Send2faTotpEmail(ctx context.Context, purpose TotpPurpose, slug, otpSecret, email string) error {
	totpCode, totpErr := totp.GenerateCodeCustom(otpSecret, time.Now(), auth.ValidateOptions)
	if totpErr != nil {
		return eris.Wrap(totpErr, "error generating custom totp code")
	}

	confBody := fmt.Sprintf(
		"Hi %s,<br/><br/>"+
			"A security code was requested for for your account. This code can be used to gain control to your account, DO NOT SHARE THIS CODE WITH OTHERS.<br/><br/>"+
			"Code: %s", slug, totpCode)

	subject := "Your Openstats Code"
	switch purpose {
	case PasswordResetPurpose:
		subject = "Your Openstats Password Reset Code"
	case EmailConfirmationPurpose:
		subject = "Your Openstats Email Confirmation Code"
	}

	return mail.Default.Send(ctx, mail.Mail{
		From:    "noreply@openstats.me",
		To:      email,
		Subject: subject,
		Body:    confBody,
	})
}

func ValidateUserEmail(ctx context.Context, userId int32, email, passcode string) (bool, error) {
	dbUserEmail, dbErr := db.Queries.GetUserEmail(ctx, userId)

	if dbErr != nil {
		return false, eris.Wrap(dbErr, "error getting user email")
	}

	if dbUserEmail.Email != email {
		return false, nil
	}

	hmacSecret, dbErr := db.Queries.SecretRead(ctx, query.SecretReadParams{
		Path: db.PrivateUser2faHmacSecretPath,
		Key:  strconv.FormatInt(int64(userId), 10),
	})

	if dbErr != nil {
		return false, eris.Wrap(dbErr, "error getting user hmac secret")
	}

	validated, validateErr := totp.ValidateCustom(passcode, hmacSecret, time.Now(), auth.ValidateOptions)
	if validateErr != nil {
		return false, eris.Wrap(validateErr, "error validating OTP")
	}

	_, dbErr = db.Queries.ConfirmEmail(ctx, query.ConfirmEmailParams{
		UserID: userId,
		Email:  email,
	})

	if dbErr != nil {
		return false, eris.Wrap(dbErr, "error confirming user email in db")
	}

	return validated, nil
}

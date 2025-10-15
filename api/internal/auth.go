package internal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dresswithpockets/openstats/app/auth"
	"github.com/dresswithpockets/openstats/app/db"
	"github.com/dresswithpockets/openstats/app/db/query"
	"github.com/dresswithpockets/openstats/app/env"
	"github.com/dresswithpockets/openstats/app/log"
	"github.com/dresswithpockets/openstats/app/password"
	"github.com/dresswithpockets/openstats/app/validation"
	"github.com/rotisserie/eris"
)

type Slug string

func (s *Slug) Schema(_ huma.Registry) *huma.Schema {
	return &huma.Schema{
		Type:               huma.TypeString,
		Title:              "Slug",
		Description:        "A human-readable identifier for a resource, such as a User or a Game",
		Pattern:            "[a-z0-9-]+",
		PatternDescription: "lowercase-alphanum with dashes",
		Format:             "slug",
		Examples:           []any{"silly-slimy-slug"},
	}
}

func (s *Slug) Resolve(ctx huma.Context, prefix *huma.PathBuffer) []error {
	if validation.ValidSlug(string(*s)) {
		return nil
	}

	return []error{&huma.ErrorDetail{
		Location: prefix.String(),
		Message:  "invalid slug",
		Value:    *s,
	}}
}

type SignInBody struct {
	Email    string `json:"email" format:"email" doc:"mutually exclusive with slug"`
	Slug     string `json:"slug" format:"slug" doc:"mutually exclusive with email" pattern:"[a-z0-9-]+" patternDescription:"lowercase-alphanum with dashes" minLength:"2" maxLength:"64"`
	Password string `json:"password" required:"true" pattern:"[a-zA-Z0-9!@#$%^&*]+" patternDescription:"alphanum with specials" minLength:"10" maxLength:"32"`
}

func (s *SignInBody) Resolve(ctx huma.Context) []error {
	if len(s.Slug) > 0 && len(s.Email) > 0 {
		return []error{&huma.ErrorDetail{
			Location: "path.slug",
			Message:  "If email is provided, slug cannot be provided",
			Value:    s.Slug,
		}}
	}

	return nil
}

var SlugMinLength int = 2
var SlugMaxLength int = 64
var PasswordMinLength int = 10
var PasswordMaxLength int = 64

type SignInInput struct {
	Body SignInBody
}

type SignInOutput struct {
	SetCookie http.Cookie `header:"Set-Cookie"`
}

func HandlePostSignIn(ctx context.Context, loginBody *SignInInput) (*SignInOutput, error) {
	result, findErr := db.Queries.FindUserBySlugWithPassword(ctx, loginBody.Body.Slug)
	if errors.Is(findErr, sql.ErrNoRows) {
		return nil, huma.Error404NotFound("slug or password don't match")
	}

	if findErr != nil {
		return nil, findErr
	}

	verifyErr := password.VerifyPassword(loginBody.Body.Password, result.EncodedHash)
	if errors.Is(verifyErr, password.ErrHashMismatch) {
		return nil, huma.Error404NotFound("slug or password don't match")
	}

	if verifyErr != nil {
		return nil, verifyErr
	}

	signedJwt, token, createErr := auth.CreateSessionToken(ctx, result.Uuid)
	if createErr != nil {
		return nil, createErr
	}

	return &SignInOutput{
		SetCookie: http.Cookie{
			Name:     auth.SessionCookieName,
			Path:     "/",
			Value:    signedJwt,
			MaxAge:   int(token.ExpiresAt.Sub(time.Now().UTC()).Seconds()),
			Expires:  token.ExpiresAt,
			Secure:   env.GetBool("OPENSTATS_SESSION_COOKIE_SECURE"),
			SameSite: http.SameSiteStrictMode,
		},
	}, nil
}

type ResetPasswordInput struct {
	Body struct {
		Slug string `json:"slug"`
		Code string `json:"code"`

		// Password is the user's login password
		Password string `json:"password" required:"true" pattern:"[a-zA-Z0-9!@#$%^&*]+" patternDescription:"alphanum with specials" minLength:"10" maxLength:"32"`
	}
}
type ResetPasswordOutput struct{}

func HandleResetPassword(ctx context.Context, input *ResetPasswordInput) (*ResetPasswordOutput, error) {
	user, err := db.Queries.FindUserBySlug(ctx, input.Body.Slug)
	if err != nil {
		return nil, huma.Error404NotFound("mismatched code or slug")
	}

	if err = auth.ReplaceUserPasswordWithTotpValidation(ctx, user.ID, input.Body.Code, input.Body.Password); err != nil {
		return nil, huma.Error404NotFound("mismatched code or slug")
	}

	return &ResetPasswordOutput{}, nil
}

type ConflictSignUpSlug struct {
	Location string
	Slug     string
}

func (c ConflictSignUpSlug) Error() string {
	return fmt.Sprintf("user slug '%s' @ '%s' is already in use", c.Slug, c.Location)
}

func (c ConflictSignUpSlug) ErrorDetail() *huma.ErrorDetail {
	return &huma.ErrorDetail{
		Message:  "the user slug is already in use",
		Location: c.Location,
		Value:    c.Slug,
	}
}

type Registration struct {
	// Email is optional, and just used for resetting the user's password
	Email string `json:"email" format:"email"`

	EmailConfirmationSent bool `json:"emailConfirmationSent" readOnly:"true" required:"false"`

	// DisplayName is optional, and is only used when displaying their profile on the website
	DisplayName *string `json:"displayName" minLength:"1" maxLength:"64" required:"false"`

	// Slug is a unique username for the user
	Slug string `json:"slug" format:"slug" required:"true" pattern:"[a-z0-9-]+" patternDescription:"lowercase-alphanum with dashes" minLength:"2" maxLength:"64"`

	// Password is the user's login password
	Password string `json:"password" required:"true" pattern:"[a-zA-Z0-9!@#$%^&*]+" patternDescription:"alphanum with specials" minLength:"10" maxLength:"32" writeOnly:"true"`
}

type SignUpRequest struct {
	Body Registration
}

type SignUpResponse struct {
	SetCookie http.Cookie `header:"Set-Cookie"`
	Body      Registration
}

func HandlePostSignUp(ctx context.Context, registerBody *SignUpRequest) (*SignUpResponse, error) {
	if auth.HasPrincipal(ctx) {
		return nil, huma.Error401Unauthorized("already signed in")
	}
	displayName := ""
	if registerBody.Body.DisplayName != nil {
		displayName = *registerBody.Body.DisplayName
	}

	createdUser, newUserError := auth.AddNewUser(
		ctx,
		displayName,
		registerBody.Body.Email,
		registerBody.Body.Slug,
		registerBody.Body.Password,
	)
	if newUserError != nil {
		if errors.Is(newUserError, db.ErrSlugAlreadyInUse) {
			return nil, &ConflictSignUpSlug{
				Location: "body.slug",
				Slug:     registerBody.Body.Slug,
			}
		}

		if validation.ErrorIsAny(newUserError, auth.ErrInvalidEmailAddress, auth.ErrInvalidDisplayName, auth.ErrInvalidSlug, auth.ErrInvalidPassword) {
			return nil, eris.Wrap(newUserError, "registerBody was Validated, and yet AddUser returned a validation error")
		}

		return nil, newUserError
	}

	emailSent := false
	if len(registerBody.Body.Email) > 0 {
		emailErr := Send2faTotpEmail(ctx, EmailConfirmationPurpose, createdUser.User.Slug, createdUser.HmacSecret, registerBody.Body.Email)
		// we log this error instead of returning it since we want a success response even if the email failed to send.
		// n.b. the user can always request the code again later
		log.Logger.Error("there was an error sending the 2FA TOTP Email", "error", emailErr)
		emailSent = emailErr == nil
	}

	signedJwt, token, createErr := auth.CreateSessionToken(ctx, createdUser.User.Uuid)
	if createErr != nil {
		return nil, createErr
	}

	return &SignUpResponse{
		SetCookie: http.Cookie{
			Name:     auth.SessionCookieName,
			Path:     "/",
			Value:    signedJwt,
			MaxAge:   int(token.ExpiresAt.Sub(time.Now().UTC()).Seconds()),
			Expires:  token.ExpiresAt,
			Secure:   env.GetBool("OPENSTATS_SESSION_COOKIE_SECURE"),
			SameSite: http.SameSiteStrictMode,
		},
		Body: Registration{
			Email:                 registerBody.Body.Email,
			EmailConfirmationSent: emailSent,
			DisplayName:           registerBody.Body.DisplayName,
			Slug:                  registerBody.Body.Slug,
		},
	}, nil
}

type SignOutResponse struct {
	SetCookie http.Cookie `header:"Set-Cookie"`
}

func HandlePostSignOut(ctx context.Context, _ *struct{}) (*SignOutResponse, error) {
	if principal, ok := auth.GetPrincipal(ctx); ok {
		err := db.Queries.DisallowToken(ctx, principal.TokenID)
		if err != nil {
			return nil, err
		}
	}

	// no matter what, always expire the session cookie
	return &SignOutResponse{
		SetCookie: http.Cookie{
			Name:     auth.SessionCookieName,
			Path:     "/",
			Value:    "",
			MaxAge:   0,
			Expires:  time.Now(),
			Secure:   env.GetBool("OPENSTATS_SESSION_COOKIE_SECURE"),
			SameSite: http.SameSiteStrictMode,
		},
	}, nil
}

type SessionResponseBody struct {
	Slug        string `json:"slug"`
	DisplayName string `json:"displayName"`
}

type SessionResponse struct {
	Body SessionResponseBody
}

func HandleGetSession(ctx context.Context, _ *struct{}) (*SessionResponse, error) {
	principal, hasPrincipal := auth.GetPrincipal(ctx)
	if !hasPrincipal {
		return nil, huma.Error401Unauthorized("no session")
	}

	userDisplayName, err := db.Queries.GetUserLatestDisplayName(ctx, principal.User.ID)
	if errors.Is(err, sql.ErrNoRows) {
		userDisplayName = query.UserDisplayName{DisplayName: ""}
	} else if err != nil {
		return nil, err
	}

	return &SessionResponse{
		Body: SessionResponseBody{
			Slug:        principal.User.Slug,
			DisplayName: userDisplayName.DisplayName,
		},
	}, nil
}

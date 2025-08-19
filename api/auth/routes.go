package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/danielgtaylor/huma/v2"
	"github.com/dresswithpockets/openstats/app/db"
	"github.com/dresswithpockets/openstats/app/db/query"
	"github.com/dresswithpockets/openstats/app/password"
	"github.com/dresswithpockets/openstats/app/validation"
	"github.com/rotisserie/eris"
	"net/http"
	"time"
)

type SignInRequest struct {
	Body struct {
		// Slug is a unique username for the user
		Slug string `json:"slug" required:"true" pattern:"[a-z0-9-]+" patternDescription:"lowercase-alphanum with dashes" minLength:"2" maxLength:"64"`

		// Password is the user's login password
		Password string `json:"password" required:"true" pattern:"[a-zA-Z0-9!@#$%^&*]+" patternDescription:"alphanum with specials" minLength:"10" maxLength:"32"`
	}
}

type SignInResponse struct {
	SetCookie http.Cookie `header:"Set-Cookie"`
}

func HandlePostSignIn(ctx context.Context, loginBody *SignInRequest) (*SignInResponse, error) {
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

	signedJwt, token, createErr := CreateSessionToken(ctx, result.Uuid)
	if createErr != nil {
		return nil, createErr
	}

	return &SignInResponse{
		SetCookie: http.Cookie{
			Name:     SessionCookieName,
			Path:     "/",
			Value:    signedJwt,
			MaxAge:   int(token.ExpiresAt.Sub(time.Now().UTC()).Seconds()),
			Expires:  token.ExpiresAt,
			Secure:   false, // TODO: enable Secure in non-local environment
			SameSite: http.SameSiteStrictMode,
		},
	}, nil
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

type SignUpBody struct {
	// Email is optional, and just used for resetting the user's password
	Email string `body:"email" format:"email"`

	// DisplayName is optional, and is only used when displaying their profile on the website
	DisplayName string `body:"displayName" minLength:"1" maxLength:"64"`

	// Slug is a unique username for the user
	Slug string `body:"slug" format:"slug" required:"true" pattern:"[a-z0-9-]+" patternDescription:"lowercase-alphanum with dashes" minLength:"2" maxLength:"64"`

	// Password is the user's login password
	Password string `body:"password" required:"true" pattern:"[a-zA-Z0-9!@#$%^&*]+" patternDescription:"alphanum with specials" minLength:"10" maxLength:"32"`
}

type SignUpRequest struct {
	Body SignUpBody
}

type SignUpResponse struct {
	SetCookie http.Cookie `header:"Set-Cookie"`
}

func HandlePostSignUp(ctx context.Context, registerBody *SignUpRequest) (*SignUpResponse, error) {
	if HasPrincipal(ctx) {
		return nil, huma.Error401Unauthorized("already signed in")
	}

	if registerBody.Body.Email != "" {
		// TODO: use SMTP email verification
	}

	newUser, newUserError := AddNewUser(
		ctx,
		registerBody.Body.DisplayName,
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

		if validation.ErrorIsAny(newUserError, ErrInvalidEmailAddress, ErrInvalidDisplayName, ErrInvalidSlug, ErrInvalidPassword) {
			return nil, eris.Wrap(newUserError, "registerBody was Validated, and yet AddUser returned a validation error")
		}

		return nil, newUserError
	}

	// TODO: send email confirmation

	signedJwt, token, createErr := CreateSessionToken(ctx, newUser.Uuid)
	if createErr != nil {
		return nil, createErr
	}

	return &SignUpResponse{
		SetCookie: http.Cookie{
			Name:     SessionCookieName,
			Path:     "/",
			Value:    signedJwt,
			MaxAge:   int(token.ExpiresAt.Sub(time.Now().UTC()).Seconds()),
			Expires:  token.ExpiresAt,
			Secure:   false, // TODO: enable Secure in non-local environment
			SameSite: http.SameSiteStrictMode,
		},
	}, nil
}

type SignOutResponse struct {
	SetCookie http.Cookie `header:"Set-Cookie"`
}

func HandlePostSignOut(ctx context.Context, input *struct{}) (*SignOutResponse, error) {
	if principal, ok := GetPrincipal(ctx); ok {
		err := db.Queries.DisallowToken(ctx, principal.TokenID)
		if err != nil {
			return nil, err
		}
	}

	// no matter what, always expire the session cookie
	return &SignOutResponse{
		SetCookie: http.Cookie{
			Name:     SessionCookieName,
			Path:     "/",
			Value:    "",
			MaxAge:   0,
			Expires:  time.Now(),
			Secure:   false, // TODO: enable Secure in non-local environment
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

func HandleGetSession(ctx context.Context, input *struct{}) (*SessionResponse, error) {
	principal, hasPrincipal := GetPrincipal(ctx)
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

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
	// Slug is a unique username for the user
	Slug string `body:"slug" required:"true" pattern:"[a-z0-9-]+" patternDescription:"lowercase-alphanum with dashes" minLength:"2" maxLength:"64"`

	// Password is the user's login password
	Password string `body:"password" required:"true" pattern:"[a-zA-Z0-9!@#$%^&*]+" patternDescription:"alphanum with specials" minLength:"10" maxLength:"32"`
}

type SignInResponse struct {
	SetCookie http.Cookie `header:"Set-Cookie"`
}

func HandlePostSignIn(ctx context.Context, loginBody *SignInRequest) (*SignInResponse, error) {
	result, findErr := db.Queries.FindUserBySlugWithPassword(ctx, loginBody.Slug)
	if errors.Is(findErr, sql.ErrNoRows) {
		return nil, huma.Error404NotFound("slug or password don't match")
	}

	if findErr != nil {
		return nil, findErr
	}

	verifyErr := password.VerifyPassword(loginBody.Password, result.EncodedHash)
	if errors.Is(verifyErr, password.ErrHashMismatch) {
		return nil, huma.Error404NotFound("slug or password don't match")
	}

	if verifyErr != nil {
		return nil, verifyErr
	}

	signedJwt, token, createErr := CreateSessionToken(ctx, result.LookupID)
	if createErr != nil {
		return nil, createErr
	}

	return &SignInResponse{
		SetCookie: http.Cookie{
			Name:    SessionCookieName,
			Value:   signedJwt,
			Expires: token.ExpiresAt,
			Secure:  true,
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

type SignUpRequest struct {
	// Email is optional, and just used for resetting the user's password
	Email string `body:"email" format:"email"`

	// DisplayName is optional, and is only used when displaying their profile on the website
	DisplayName string `body:"displayName" minLength:"1" maxLength:"64"`

	// Slug is a unique username for the user
	Slug string `body:"slug" format:"slug" required:"true" pattern:"[a-z0-9-]+" patternDescription:"lowercase-alphanum with dashes" minLength:"2" maxLength:"64"`

	// Password is the user's login password
	Password string `body:"password" required:"true" pattern:"[a-zA-Z0-9!@#$%^&*]+" patternDescription:"alphanum with specials" minLength:"10" maxLength:"32"`
}

type SignUpResponse struct {
	SetCookie http.Cookie `header:"Set-Cookie"`
}

func HandlePostSignUp(ctx context.Context, registerBody *SignUpRequest) (*SignUpResponse, error) {
	if HasPrincipal(ctx) {
		return nil, huma.Error401Unauthorized("already signed in")
	}

	newUser, newUserError := AddNewUser(
		ctx,
		registerBody.DisplayName,
		registerBody.Email,
		registerBody.Slug,
		registerBody.Password,
	)
	if newUserError != nil {
		if errors.Is(newUserError, db.ErrSlugAlreadyInUse) {
			return nil, &ConflictSignUpSlug{
				Location: "body.slug",
				Slug:     registerBody.Slug,
			}
		}

		if validation.ErrorIsAny(newUserError, ErrInvalidEmailAddress, ErrInvalidDisplayName, ErrInvalidSlug, ErrInvalidPassword) {
			return nil, eris.Wrap(newUserError, "registerBody was Validated, and yet AddUser returned a validation error")
		}

		return nil, newUserError
	}

	signedJwt, token, createErr := CreateSessionToken(ctx, newUser.LookupID)
	if createErr != nil {
		return nil, createErr
	}

	return &SignUpResponse{
		SetCookie: http.Cookie{
			Name:    SessionCookieName,
			Value:   signedJwt,
			Expires: token.ExpiresAt,
			Secure:  true,
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
			Name:    SessionCookieName,
			Expires: time.Now(),
			Secure:  true,
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

func RegisterRoutes(api huma.API) {
	authApi := huma.NewGroup(api, "/auth/v1")
	authApi.UseSimpleModifier(func(op *huma.Operation) {
		op.Tags = append(op.Tags, "Authentication v1")
	})
	authApi.UseMiddleware(UserAuthHandler)

	requireUserAuthHandler := CreateRequireUserAuthHandler(authApi)

	huma.Register(authApi, huma.Operation{
		OperationID: "get-session",
		Method:      http.MethodGet,
		Path:        "/session",
		Summary:     "Get session info",
		Description: "Get the current authenticated session's user info",
		Security:    []map[string][]string{{"Cookie": {}}},
		Middlewares: huma.Middlewares{requireUserAuthHandler},
		Errors: []int{
			http.StatusUnauthorized,
		},
	}, HandleGetSession)

	huma.Register(authApi, huma.Operation{
		OperationID: "sign-in",
		Method:      http.MethodPost,
		Path:        "/sign-in",
		Summary:     "Sign In",
		Description: "Sign into a new session",
		Errors: []int{
			http.StatusUnauthorized,
			http.StatusBadRequest,
		},
	}, HandlePostSignIn)

	huma.Register(authApi, huma.Operation{
		OperationID: "sign-up",
		Method:      http.MethodPost,
		Path:        "/sign-up",
		Summary:     "Sign Up",
		Description: "Create a new user and sign into a new session as the new user",
		Errors: []int{
			http.StatusUnauthorized,
			http.StatusBadRequest,
		},
	}, HandlePostSignUp)

	huma.Register(authApi, huma.Operation{
		OperationID: "sign-out",
		Method:      http.MethodPost,
		Path:        "/sign-out",
		Summary:     "Sign Out",
		Description: "Sign out of the current session, and invalidate the session token",
		Security:    []map[string][]string{{"Cookie": {}}},
		Middlewares: huma.Middlewares{requireUserAuthHandler},
		Errors: []int{
			http.StatusUnauthorized,
		},
	}, HandlePostSignOut)
}

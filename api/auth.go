package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/danielgtaylor/huma/v2"
	"github.com/dresswithpockets/openstats/app/db"
	"github.com/dresswithpockets/openstats/app/db/query"
	"github.com/dresswithpockets/openstats/app/password"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"log"
	"net/http"
	"time"
)

var ArgonParameters = password.Parameters{
	Iterations:  2,
	Memory:      19 * 1024,
	Parallelism: 1,
	SaltLength:  16,
	KeyLength:   32,
}

const SessionCookieName = "sessionid"
const SessionIssuer = "openstats"
const SessionAudience = "openstats"
const SessionDuration = time.Hour * 24 * 7
const SessionJitter = time.Minute
const PrincipalContextKey = "principal"

var SessionTokenSecret = []byte("blahblahblah")

type Principal struct {
	User    query.User
	TokenID uuid.UUID
	Claims  jwt.RegisteredClaims
}

func GetPrincipal(ctx context.Context) (result *Principal, ok bool) {
	result, ok = ctx.Value(PrincipalContextKey).(*Principal)
	ok = ok && result != nil
	return
}

func HasPrincipal(ctx context.Context) bool {
	result, ok := ctx.Value(PrincipalContextKey).(*Principal)
	return ok && result != nil
}

func UserAuthHandler(ctx huma.Context, next func(huma.Context)) {
	sessionCookie, cookieErr := huma.ReadCookie(ctx, SessionCookieName)
	if cookieErr != nil {
		next(ctx)
		return
	}

	token, parseErr := jwt.ParseWithClaims(sessionCookie.Value, jwt.RegisteredClaims{}, func(token *jwt.Token) (any, error) {
		return SessionTokenSecret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithIssuer(SessionIssuer), jwt.WithAudience(SessionAudience))
	if parseErr != nil {
		next(ctx)
		return
	}

	// the subject should be a user lookup ulid which never changes per-user
	subject, subjectErr := token.Claims.GetSubject()
	if subjectErr != nil {
		next(ctx)
		return
	}

	subjectUuid, uuidErr := uuid.Parse(subject)
	if uuidErr != nil {
		log.Println(uuidErr)
		next(ctx)
		return
	}

	sessionUser, findErr := Queries.FindUserByLookupId(ctx.Context(), subjectUuid)
	if errors.Is(findErr, sql.ErrNoRows) {
		next(ctx)
		return
	}

	if findErr != nil {
		// TODO: huma.WriteErr() return as problem details? I really wish middleware could go through our error handler...
		next(ctx)
		return
	}

	tokenId, tokenIdErr := uuid.Parse(token.Claims.(jwt.RegisteredClaims).ID)
	if tokenIdErr != nil {
		// TODO: huma.WriteErr(), the JTI should always be a UUID...
		next(ctx)
		return
	}

	ctx = huma.WithValue(ctx, PrincipalContextKey, &Principal{
		User:    sessionUser,
		TokenID: tokenId,
		Claims:  token.Claims.(jwt.RegisteredClaims),
	})
	next(ctx)
}

type RequireUserAuthMiddleware struct {
	api huma.API
}

func NewRequireUserAuthMiddleware(api huma.API) *RequireUserAuthMiddleware {
	return &RequireUserAuthMiddleware{api: api}
}

func (m RequireUserAuthMiddleware) Handler(ctx huma.Context, next func(huma.Context)) {
	if !HasPrincipal(ctx.Context()) {
		_ = huma.WriteErr(m.api, ctx, http.StatusUnauthorized, "")
		return
	}

	next(ctx)
}

type RequireAdminAuthMiddleware struct {
	api huma.API
}

func NewRequireAdminAuthMiddleware(api huma.API) *RequireAdminAuthMiddleware {
	return &RequireAdminAuthMiddleware{api: api}
}

func (m RequireAdminAuthMiddleware) Handler(ctx huma.Context, next func(huma.Context)) {
	// TODO: include role in JWT?
	if principal, ok := GetPrincipal(ctx.Context()); !ok || !IsAdmin(principal.User) {
		_ = huma.WriteErr(m.api, ctx, http.StatusUnauthorized, "")
	}

	next(ctx)
}

var (
	ErrInvalidEmailAddress = errors.New("invalid email address")
	ErrInvalidDisplayName  = errors.New("invalid display name")
	ErrInvalidSlug         = errors.New("invalid slug")
	ErrInvalidPassword     = errors.New("invalid password")
)

func AddNewUser(ctx context.Context, displayName, email, slug, pass string) (newUser *query.User, err error) {
	if len(email) > 0 && !ValidEmailAddress(email) {
		return nil, eris.Wrap(ErrInvalidEmailAddress, "validation error")
	}

	if len(displayName) > 0 && !ValidDisplayName(displayName) {
		return nil, eris.Wrap(ErrInvalidDisplayName, "validation error")
	}

	if !ValidSlug(slug) {
		return nil, eris.Wrap(ErrInvalidSlug, "validation error")
	}

	if !ValidPassword(pass) {
		return nil, eris.Wrap(ErrInvalidPassword, "validation error")
	}

	encodedPassword, passwordErr := password.EncodePassword(pass, ArgonParameters)
	if passwordErr != nil {
		return nil, passwordErr
	}

	return Actions.CreateUser(ctx, slug, encodedPassword, email, displayName)
}

func CreateSessionToken(ctx context.Context, userLookupId uuid.UUID) (signedToken string, token query.Token, err error) {
	nowTime := time.Now().UTC()
	token, err = Queries.CreateToken(ctx, query.CreateTokenParams{
		Issuer:    SessionIssuer,
		Subject:   userLookupId.String(),
		Audience:  SessionAudience,
		ExpiresAt: nowTime.Add(SessionDuration),
		NotBefore: nowTime.Add(-SessionJitter),
		IssuedAt:  nowTime,
	})
	if err != nil {
		return
	}

	claims := jwt.RegisteredClaims{
		Issuer:    token.Issuer,
		Subject:   token.Subject,
		Audience:  []string{token.Audience},
		ExpiresAt: jwt.NewNumericDate(token.ExpiresAt),
		NotBefore: jwt.NewNumericDate(token.NotBefore),
		IssuedAt:  jwt.NewNumericDate(token.IssuedAt),
		ID:        token.ID.String(),
	}

	signedToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(SessionTokenSecret)
	if err != nil {
		return
	}

	return
}

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
	result, findErr := Queries.FindUserBySlugWithPassword(ctx, loginBody.Slug)
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

		if ErrorIsAny(newUserError, ErrInvalidEmailAddress, ErrInvalidDisplayName, ErrInvalidSlug, ErrInvalidPassword) {
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
		err := Queries.DisallowToken(ctx, principal.TokenID)
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

	userDisplayName, err := Queries.GetUserLatestDisplayName(ctx, principal.User.ID)
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

package auth

import (
	"context"
	"errors"
	"github.com/dresswithpockets/openstats/app/db"
	"github.com/dresswithpockets/openstats/app/db/query"
	"github.com/dresswithpockets/openstats/app/password"
	"github.com/dresswithpockets/openstats/app/rid"
	"github.com/dresswithpockets/openstats/app/validation"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"time"
)

var ArgonParameters = password.Parameters{
	Iterations:  2,
	Memory:      19 * 1024,
	Parallelism: 1,
	SaltLength:  16,
	KeyLength:   32,
}

const (
	SessionCookieName   = "sessionid"
	SessionIssuer       = "openstats"
	SessionAudience     = "openstats"
	SessionDuration     = time.Hour * 24 * 7
	SessionJitter       = time.Minute
	PrincipalContextKey = "principal"

	GameSessionIssuer   = "openstats"
	GameSessionAudience = "openstats"
)

var SessionTokenSecret = []byte("blahblahblah")

type Principal struct {
	User    query.User
	TokenID uuid.UUID
	Claims  *jwt.RegisteredClaims
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

func GetGameTokenPrincipal(ctx context.Context) (result *GameTokenPrincipal, ok bool) {
	result, ok = ctx.Value(PrincipalContextKey).(*GameTokenPrincipal)
	ok = ok && result != nil
	return
}

func HasGameTokenPrincipal(ctx context.Context) bool {
	result, ok := ctx.Value(PrincipalContextKey).(*GameTokenPrincipal)
	return ok && result != nil
}

func GetGameSessionPrincipal(ctx context.Context) (result *GameSessionPrincipal, ok bool) {
	result, ok = ctx.Value(PrincipalContextKey).(*GameSessionPrincipal)
	ok = ok && result != nil
	return
}
func HasGameSessionPrincipal(ctx context.Context) bool {
	result, ok := ctx.Value(PrincipalContextKey).(*GameSessionPrincipal)
	return ok && result != nil
}

var (
	ErrInvalidEmailAddress = errors.New("invalid email address")
	ErrInvalidDisplayName  = errors.New("invalid display name")
	ErrInvalidSlug         = errors.New("invalid slug")
	ErrInvalidPassword     = errors.New("invalid password")
)

func AddNewUser(ctx context.Context, displayName, email, slug, pass string) (newUser *query.User, err error) {
	if len(email) > 0 && !validation.ValidEmailAddress(email) {
		return nil, eris.Wrap(ErrInvalidEmailAddress, "validation error")
	}

	if len(displayName) > 0 && !validation.ValidDisplayName(displayName) {
		return nil, eris.Wrap(ErrInvalidDisplayName, "validation error")
	}

	if !validation.ValidSlug(slug) {
		return nil, eris.Wrap(ErrInvalidSlug, "validation error")
	}

	if !validation.ValidPassword(pass) {
		return nil, eris.Wrap(ErrInvalidPassword, "validation error")
	}

	encodedPassword, passwordErr := password.EncodePassword(pass, ArgonParameters)
	if passwordErr != nil {
		return nil, passwordErr
	}

	return db.DB.CreateUser(ctx, slug, encodedPassword, email, displayName)
}

func CreateSessionToken(ctx context.Context, userLookupId uuid.UUID) (signedToken string, token query.Token, err error) {
	nowTime := time.Now().UTC()
	token, err = db.Queries.CreateToken(ctx, query.CreateTokenParams{
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

func CreateGameSessionToken(ctx context.Context, gameTokenUuid uuid.UUID, userRid, gameRid rid.RID) (signedToken string, gameSession query.GameSession, err error) {
	var token query.Token
	token, gameSession, err = db.DB.CreateGameSessionAndToken(
		ctx,
		gameTokenUuid,
		userRid,
		gameRid,
		GameSessionIssuer,
		GameSessionAudience,
		time.Hour,
		SessionJitter,
	)
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

package auth

import (
	"context"
	"database/sql"
	"errors"
	"github.com/danielgtaylor/huma/v2"
	"github.com/dresswithpockets/openstats/app/db"
	"github.com/dresswithpockets/openstats/app/db/query"
	"github.com/dresswithpockets/openstats/app/rid"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"log"
	"net/http"
	"strings"
	"time"
)

const UserRidPrefix = "u"
const GameRidPrefix = "g"
const GameSessionRidPrefix = "gs"

type GameSessionPrincipal struct {
	SessionRid    rid.RID
	UserRid       rid.RID
	GameRid       rid.RID
	GameTokenUuid uuid.UUID
	LastPulse     time.Time
	ExpiresAt     time.Time
}

var ErrInvalidGameSessionToken = errors.New("invalid game session token")

func ensureGameSessionPrincipal(ctx context.Context, claims *jwt.RegisteredClaims) (principal *GameSessionPrincipal, err error) {
	// at the moment there is only one format for Subject
	// Subject identifies the authorized user, in format `users/v1/{userRID}`

	subjectParts := strings.Split(claims.Subject, "/")
	if len(subjectParts) != 7 || subjectParts[0] != "users" || subjectParts[1] != "v1" || subjectParts[3] != "games" || subjectParts[5] != "sessions" {
		return nil, ErrInvalidGameSessionToken
	}

	userRid, userRidErr := rid.ParseString(subjectParts[2])
	if userRidErr != nil || userRid.Prefix != UserRidPrefix {
		return nil, ErrInvalidGameSessionToken
	}

	gameRid, gameRidErr := rid.ParseString(subjectParts[4])
	if gameRidErr != nil {
		return nil, ErrInvalidGameSessionToken
	}

	sessionRid, sessionRidErr := rid.ParseString(subjectParts[6])
	if sessionRidErr != nil {
		return nil, ErrInvalidGameSessionToken
	}

	tokenId, tokenIdErr := uuid.Parse(claims.ID)
	if tokenIdErr != nil {
		return nil, ErrInvalidGameSessionToken
	}

	result, dbErr := db.Queries.GetValidSession(ctx, query.GetValidSessionParams{
		SessionTokenUuid: tokenId,
		GameUuid:         gameRid.ID,
		UserUuid:         userRid.ID,
		SessionUuid:      sessionRid.ID,
	})
	if errors.Is(dbErr, sql.ErrNoRows) {
		return nil, ErrInvalidGameSessionToken
	}

	if dbErr != nil {
		return nil, dbErr
	}

	return &GameSessionPrincipal{
		SessionRid:    sessionRid,
		UserRid:       userRid,
		GameRid:       gameRid,
		GameTokenUuid: result.GameTokenUuid,
		LastPulse:     result.LastPulseAt,
		ExpiresAt:     claims.ExpiresAt.Time,
	}, nil
}

//func ensureGameSessionPrincipalOld(ctx context.Context, claims *jwt.RegisteredClaims) (principal *GameSessionPrincipal, err error) {
//	// at the moment there is only one format for Issuer, Subject, and Audience
//	// Subject identifies the authorized user, in format `users/v1/{userRID}`
//	userRidString := strings.TrimPrefix(claims.Subject, "users/v1/")
//	if userRidString == "" || userRidString == claims.Subject {
//		return nil, ErrInvalidGameSessionToken
//	}
//
//	userRid, userRidErr := rid.ParseString(userRidString)
//	if userRidErr != nil || userRid.Prefix != UserRidPrefix {
//		return nil, ErrInvalidGameSessionToken
//	}
//
//	if !strings.HasPrefix(claims.Issuer, "users/v1/") || !strings.HasSuffix(claims.Issuer, "/session") {
//		return nil, ErrInvalidGameSessionToken
//	}
//
//	// Issuer identifies where the token was issued, in the format `users/v1/{userRID}/sessions`
//	expectedIssuerString := fmt.Sprintf("users/v1/%s/sessions", userRid.String())
//	if claims.Issuer != expectedIssuerString {
//		return nil, ErrInvalidGameSessionToken
//	}
//
//	// we only produce tokens with a single audience at the moment
//	if len(claims.Audience) != 1 {
//		return nil, ErrInvalidGameSessionToken
//	}
//
//	audience := claims.Audience[0]
//
//	// Audience identifies the specific game that this session token is authorized for
//	// in format `developers/v1/{developerRID}/games/{gameRID}`
//	var gamesPathStartIndex = strings.Index(audience, "/games/")
//	var gamePathEndIndex = gamesPathStartIndex + len("/games/")
//	developerRidString := audience[len("developers/v1/"):gamesPathStartIndex]
//	gameRidString := audience[gamePathEndIndex:]
//
//	developerRid, developerRidErr := rid.ParseString(developerRidString)
//	if developerRidErr != nil || developerRid.Prefix != "d" {
//		return nil, ErrInvalidGameSessionToken
//	}
//
//	gameRid, gameRidErr := rid.ParseString(gameRidString)
//	if gameRidErr != nil || gameRid.Prefix != "g" {
//		return nil, ErrInvalidGameSessionToken
//	}
//
//	result, dbErr := db.Queries.GetGameSessionRidCounts(ctx, query.GetGameSessionRidCountsParams{
//		UserUuid:      userRid.ID,
//		DeveloperUuid: developerRid.ID,
//		GameUuid:      gameRid.ID,
//	})
//	if dbErr != nil {
//		return nil, dbErr
//	}
//
//	// the game and user must exist, and the JWT must not be in the disallow list
//	if result.GameCount != 1 || result.UserCount != 1 || result.DisallowCount != 0 {
//		return nil, ErrInvalidGameSessionToken
//	}
//
//	return &GameSessionPrincipal{
//		UserUuid:      userRid.ID,
//		DeveloperUuid: developerRid.ID,
//		GameUuid:      gameRid.ID,
//	}, nil
//}

func GameSessionAuthHandler(ctx huma.Context, next func(huma.Context)) {
	authHeader := ctx.Header("Authorization")
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == "" || tokenString == authHeader {
		next(ctx)
		return
	}

	/*
		game session tokens are JWTs generated when a game session is created, with these claims:

		sub: the path to the authorized session e.g. `users/v1/{userRID}/games/{gameRID}/sessions/{sessionRID}`
		exp: the token's expiration timestamp, which is always iat + a duration chosen by the session creator
		nbf: always the timestamp that the token was created at
		iat: always the timestamp that the token was created at
		jti: a unique identifier for the JWT, unique across all openstats JWTs

		the claims are used to verify that the submitter has permission to submit achievement progress and
		game stats for a particular user.

		the token table itself just stores information about issued tokens, it is not used for claims validation,
		authentication, or authorization. Only the private key & JWT claims are used for those.
	*/

	token, parseErr := jwt.ParseWithClaims(
		tokenString,
		&jwt.RegisteredClaims{},
		func(token *jwt.Token) (any, error) { return SessionTokenSecret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
		jwt.WithIssuer(GameSessionIssuer),
		jwt.WithAudience(GameSessionAudience),
	)
	if parseErr != nil {
		next(ctx)
		return
	}

	gameSessionClaims, isRegisteredClaims := token.Claims.(*jwt.RegisteredClaims)
	if !isRegisteredClaims {
		next(ctx)
		return
	}

	principal, ensureErr := ensureGameSessionPrincipal(ctx.Context(), gameSessionClaims)
	if ensureErr != nil {
		next(ctx)
		return
	}

	ctx = huma.WithValue(ctx, PrincipalContextKey, principal)
	next(ctx)
}

type GameTokenPrincipal struct {
	TokenUuid uuid.UUID
	UserRid   rid.RID
	GameRid   rid.RID
}

func GameTokenAuthHandler(ctx huma.Context, next func(huma.Context)) {
	authHeader := ctx.Header("Authorization")
	if authHeader == "" {
		next(ctx)
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == "" || tokenString == authHeader {
		next(ctx)
		return
	}

	uuidBytes, decodeErr := rid.Base62Encoding.Decode(tokenString)
	if decodeErr != nil {
		next(ctx)
		return
	}

	tokenUuid := uuid.UUID(uuidBytes)
	if tokenUuid == uuid.Nil {
		next(ctx)
		return
	}

	tokenInfo, findErr := db.Queries.FindTokenWithUser(ctx.Context(), tokenUuid)
	if findErr != nil {
		next(ctx)
		return
	}

	// TODO: differentiate between a User Identity/Principal and a GameToken Identity/Principal
	ctx = huma.WithValue(ctx, PrincipalContextKey, &GameTokenPrincipal{
		TokenUuid: tokenUuid,
		UserRid:   rid.From(UserRidPrefix, tokenInfo.UserUuid),
		GameRid:   rid.From(GameRidPrefix, tokenInfo.GameUuid),
	})
	next(ctx)
}

func UserAuthHandler(ctx huma.Context, next func(huma.Context)) {
	sessionCookie, cookieErr := huma.ReadCookie(ctx, SessionCookieName)
	if cookieErr != nil {
		next(ctx)
		return
	}

	token, parseErr := jwt.ParseWithClaims(sessionCookie.Value, &jwt.RegisteredClaims{}, func(token *jwt.Token) (any, error) {
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

	sessionUser, findErr := db.Queries.FindUser(ctx.Context(), subjectUuid)
	if errors.Is(findErr, sql.ErrNoRows) {
		next(ctx)
		return
	}

	if findErr != nil {
		// TODO: huma.WriteErr() return as problem details? I really wish middleware could go through our error handler...
		next(ctx)
		return
	}

	tokenId, tokenIdErr := uuid.Parse(token.Claims.(*jwt.RegisteredClaims).ID)
	if tokenIdErr != nil {
		// TODO: huma.WriteErr(), the JTI should always be a UUID...
		next(ctx)
		return
	}

	// TODO: the tokenId shouldn't be in the disallow list...

	ctx = huma.WithValue(ctx, PrincipalContextKey, &Principal{
		User:    sessionUser,
		TokenID: tokenId,
		Claims:  token.Claims.(*jwt.RegisteredClaims),
	})
	next(ctx)
}

func CreateRequireUserAuthHandler(api huma.API) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		if !HasPrincipal(ctx.Context()) {
			_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, "")
			return
		}

		next(ctx)
	}
}

func CreateRequireGameTokenAuthHandler(api huma.API) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		if !HasGameTokenPrincipal(ctx.Context()) {
			_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, "")
			return
		}

		next(ctx)
	}
}

func CreateRequireGameSessionAuthHandler(api huma.API) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		if !HasGameSessionPrincipal(ctx.Context()) {
			_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, "")
			return
		}

		next(ctx)
	}
}

func CreateRequireAdminAuthHandler(api huma.API) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		// TODO: include role in JWT?
		if principal, ok := GetPrincipal(ctx.Context()); !ok || !IsAdmin(principal.User) {
			_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, "")
		}

		next(ctx)
	}
}

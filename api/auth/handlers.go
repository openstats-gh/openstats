package auth

import (
	"database/sql"
	"errors"
	"github.com/danielgtaylor/huma/v2"
	"github.com/dresswithpockets/openstats/app/db"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"log"
	"net/http"
)

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

	sessionUser, findErr := db.Queries.FindUserByLookupId(ctx.Context(), subjectUuid)
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

func CreateRequireUserAuthHandler(api huma.API) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		if !HasPrincipal(ctx.Context()) {
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

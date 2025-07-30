package main

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"

	"github.com/rs/cors"
)

func HandlerTODO(ctx context.Context, input *struct{}) (*struct{}, error) {
	panic("HandlerTODO")
}

func main() {
	if err := SetupDB(context.Background()); err != nil {
		log.Fatal(err)
	}

	if err := SetupValidations(); err != nil {
		log.Fatal(err)
	}

	// we need a root admin user in order to do admin operations. The root user is also the only user that can add
	// other admins
	AddRootAdminUser(context.Background())

	router := chi.NewMux()
	router.Use(cors.New(cors.Options{
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowCredentials: true,
	}).Handler)
	// TODO: CSRF middleware
	// TODO: rate limit middleware

	config := huma.DefaultConfig("openstats API", "1.0.0")
	config.Info = &huma.Info{
		Title:       "openstats API",
		Description: "The public openstats API",
		Contact: &huma.Contact{
			Name: "dresswithpockets",
			URL:  "https://github.com/dresswithpockets",
		},
		License: &huma.License{
			Name:       "GPL General Public License v3",
			Identifier: "GPL-3.0-or-later",
			URL:        "https://spdx.org/licenses/GPL-3.0-or-later.html",
		},
		Version: "v1.0.0",
	}
	config.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"Cookie": {
			Type:        "apiKey",
			In:          "cookie",
			Description: "A basic user authentication, typically created by sign-up or sign-in",
			Name:        SessionCookieName,
		},
	}
	api := humachi.New(router, config)

	requireUserAuthHandler := NewRequireUserAuthMiddleware(api)

	type ReadyResponse struct{ OK bool }
	huma.Register(api, huma.Operation{
		OperationID: "readyz",
		Method:      http.MethodGet,
		Path:        "/readyz",
		Summary:     "Get Readiness",
		Description: "Get whether or not the API is ready to process requests",
		Tags:        []string{"Health Check"},
	}, func(ctx context.Context, _ *struct{}) (*ReadyResponse, error) {
		return &ReadyResponse{OK: true}, nil
	})

	authApi := huma.NewGroup(api, "/auth/v1")
	authApi.UseSimpleModifier(func(op *huma.Operation) {
		op.Tags = append(op.Tags, "Authentication v1")
	})
	authApi.UseMiddleware(UserAuthHandler)

	huma.Register(authApi, huma.Operation{
		OperationID: "get-session",
		Method:      http.MethodGet,
		Path:        "/session",
		Summary:     "Get session info",
		Description: "Get the current authenticated session's user info",
		Security:    []map[string][]string{{"Cookie": {}}},
		Middlewares: huma.Middlewares{requireUserAuthHandler.Handler},
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
		Middlewares: huma.Middlewares{requireUserAuthHandler.Handler},
		Errors: []int{
			http.StatusUnauthorized,
		},
	}, HandlePostSignOut)

	userApi := huma.NewGroup(api, "/users/v1")
	userApi.UseSimpleModifier(func(op *huma.Operation) {
		op.Tags = append(op.Tags, "Users")
	})
	userApi.UseMiddleware(UserAuthHandler)

	huma.Register(userApi, huma.Operation{
		OperationID: "get-users-brief",
		Method:      http.MethodGet,
		Path:        "/{slug}/brief",
		Summary:     "Get user brief",
		Description: "Get a detail summary containing the user's recent achievements, for display",
	}, HandleGetUsersBrief)

	huma.Register(userApi, huma.Operation{
		OperationID: "list-users",
		Method:      http.MethodGet,
		Path:        "/",
		Summary:     "List users",
		Description: "Query & filter all users",
	}, HandlerTODO)

	huma.Register(userApi, huma.Operation{
		OperationID: "create-users",
		Method:      http.MethodPost,
		Path:        "/",
		Summary:     "Create new users",
		Description: "Create 1 or more users. Requires an admin session.",
		Security:    []map[string][]string{{"Cookie": {}}},
		Middlewares: huma.Middlewares{requireUserAuthHandler.Handler},
		Errors: []int{
			http.StatusUnauthorized,
		},
	}, HandlerTODO)

	huma.Register(userApi, huma.Operation{
		OperationID: "read-user",
		Method:      http.MethodGet,
		Path:        "/{slug}",
		Summary:     "Read user",
		Description: "Get some details for a particular user",
	}, HandlerTODO)

	huma.Register(userApi, huma.Operation{
		OperationID: "upsert-user",
		Method:      http.MethodPut,
		Path:        "/{slug}",
		Summary:     "Create or update user",
		Description: "Create or update a user at the slug specified. This is an upsert operation - it will try to create the user if it doesn't already exist, and will otherwise update an existing user.",
		Security:    []map[string][]string{{"Cookie": {}}},
		Middlewares: huma.Middlewares{requireUserAuthHandler.Handler},
		Errors: []int{
			http.StatusUnauthorized,
		},
	}, HandlerTODO)

	huma.Register(userApi, huma.Operation{
		OperationID: "patch-user",
		Method:      http.MethodPatch,
		Path:        "/{slug}",
		Summary:     "Update user",
		Description: "Update an existing user at the slug.",
		Security:    []map[string][]string{{"Cookie": {}}},
		Middlewares: huma.Middlewares{requireUserAuthHandler.Handler},
		Errors: []int{
			http.StatusUnauthorized,
		},
	}, HandlerTODO)

	huma.Register(userApi, huma.Operation{
		OperationID: "delete-user",
		Method:      http.MethodDelete,
		Path:        "/{slug}",
		Summary:     "Delete a user",
		Description: "Delete an existing user at the slug. Must be an Admin.",
		Security:    []map[string][]string{{"Cookie": {}}},
		Middlewares: huma.Middlewares{requireUserAuthHandler.Handler},
		Errors: []int{
			http.StatusUnauthorized,
		},
	}, HandlerTODO)

	if err := http.ListenAndServe(":3000", router); err != nil {
		log.Fatal(err)
	}

	//ErrorHandler: func(c *fiber.Ctx, err error) error {
	//	var validationErr *ValidationError
	//	if errors.As(err, &validationErr) {
	//		var fieldErrors map[string][]string
	//		for _, fieldError := range validationErr.Errors {
	//			detail := GetValidationDetail(fieldError.Field)
	//			fieldErrors[fieldError.Field] = append(fieldErrors[fieldError.Field], detail)
	//		}
	//
	//		c.Status(fiber.StatusBadRequest)
	//		return c.JSON(problems.Validation("", fieldErrors))
	//	}
	//
	//	var conflictErr *ConflictError
	//	if errors.As(err, &conflictErr) {
	//		c.Status(fiber.StatusConflict)
	//		return c.JSON(problems.Conflict(conflictErr.Field, conflictErr.Value, ""))
	//	}
	//
	//	// TODO: request IDs to associate with errors
	//	// TODO: setup default logger to output in a queryable format e.g JSON
	//	log.Error(err)
	//
	//	// TODO: replace err with non-descriptive "An error occurred on the server" in production
	//	return fiber.DefaultErrorHandler(c, err)
	//},
}

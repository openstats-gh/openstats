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
	config.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"Cookie": {
			Type: "apiKey",
			In:   "cookie",
			Name: SessionCookieName,
		},
	}
	api := humachi.New(router, config)

	requireUserAuthHandler := NewRequireUserAuthMiddleware(api)
	// TODO: Authentication middleware
	// TODO: Authorization middleware

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
	authApi.UseModifier(func(op *huma.Operation, next func(*huma.Operation)) {
		op.Tags = append(op.Tags, "Authentication v1")
		next(op)
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
	}, HandleGetSession)

	huma.Register(authApi, huma.Operation{
		OperationID: "sign-in",
		Method:      http.MethodPost,
		Path:        "/sign-in",
		Summary:     "Sign In",
		Description: "Sign into a new session",
	}, HandlePostSignIn)

	huma.Register(authApi, huma.Operation{
		OperationID: "sign-up",
		Method:      http.MethodPost,
		Path:        "/sign-up",
		Summary:     "Sign Up",
		Description: "Create a new user and sign into a new session as the new user",
	}, HandlePostSignUp)

	huma.Register(authApi, huma.Operation{
		OperationID: "sign-out",
		Method:      http.MethodPost,
		Path:        "/sign-out",
		Summary:     "Sign Out",
		Description: "Sign out of the current session, and invalidate the session token",
		Security:    []map[string][]string{{"Cookie": {}}},
		Middlewares: huma.Middlewares{requireUserAuthHandler.Handler},
	}, HandlePostSignOut)

	if err := http.ListenAndServe(":3000", router); err != nil {
		log.Fatal(err)
	}

	//server := fuego.NewServer(
	//	fuego.WithGlobalMiddlewares(cors.New(cors.Options{
	//		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	//		AllowCredentials: true,
	//	}).Handler),
	//)
	//
	//fuego.Get(server, "/readyz", func(c fuego.ContextNoBody) (string, error) {
	//	return "OK", nil
	//})
	//
	//authGroup := fuego.Group(server, "auth/v1", option.Middleware(UserAuthHandler))
	//fuego.Get(authGroup, "/session", HandleGetSession)
	//fuego.Post(authGroup, "/sign-in", HandlePostSignIn)
	//fuego.Post(authGroup, "/sign-up", HandlePostSignUp, option.Middleware(RequireUserAuthHandler))
	//fuego.Post(authGroup, "/sign-out", HandlePostSignOut)
	//
	//// TODO: user or profile apis
	//// TODO: developer apis
	//// TODO: developer-game apis
	//// TODO: developer-game-achievement apis
	//
	//if err := server.Run(); err != nil {
	//	log.Fatal(err)
	//}

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

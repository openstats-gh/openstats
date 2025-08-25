package main

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/dresswithpockets/openstats/app/auth"
	"github.com/dresswithpockets/openstats/app/db"
	"github.com/dresswithpockets/openstats/app/env"
	"github.com/dresswithpockets/openstats/app/internal"
	"github.com/dresswithpockets/openstats/app/log"
	"github.com/dresswithpockets/openstats/app/mail"
	"github.com/dresswithpockets/openstats/app/media"
	"github.com/dresswithpockets/openstats/app/users"
	"github.com/dresswithpockets/openstats/app/validation"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v3"
	"github.com/rotisserie/eris"
	"github.com/rs/cors"
	golog "log"
	"log/slog"
	"net/http"
	"os"
)

func setupRouter() (*chi.Mux, error) {
	logConcise := env.GetBool("OPENSTATS_HTTPLOG_CONCISE")
	logFormat := httplog.SchemaECS.Concise(logConcise)

	logLevel, matchedErr := env.GetMapped("OPENSTATS_HTTPLOG_LEVEL", log.SlogLevelMap)
	if matchedErr != nil {
		return nil, matchedErr
	}

	handlerOptions := &slog.HandlerOptions{
		ReplaceAttr: logFormat.ReplaceAttr,
		Level:       logLevel,
	}

	var logger *slog.Logger
	slogMode := env.GetString("OPENSTATS_HTTPLOG_MODE")
	switch slogMode {
	case "Text":
		logger = slog.New(slog.NewTextHandler(os.Stdout, handlerOptions))
	case "JSON":
		logger = slog.New(slog.NewJSONHandler(os.Stdout, handlerOptions))
	default:
		return nil, eris.Errorf("invalid value for OPENSTATS_HTTPLOG_Mode: %s", slogMode)
	}

	router := chi.NewMux()

	options := &httplog.Options{
		// Level defines the verbosity of the request logs:
		// slog.LevelDebug - log all responses (incl. OPTIONS)
		// slog.LevelInfo  - log responses (excl. OPTIONS)
		// slog.LevelWarn  - log 4xx and 5xx responses only (except for 429)
		// slog.LevelError - log 5xx responses only
		Level: logLevel,

		// Set log output to Elastic Common Schema (ECS) format.
		Schema: logFormat,

		// RecoverPanics recovers from panics occurring in the underlying HTTP handlers
		// and middlewares. It returns HTTP 500 unless response status was already set.
		//
		// NOTE: Panics are logged as errors automatically, regardless of this setting.
		RecoverPanics: true,

		// Optionally, log selected request/response headers explicitly.
		LogRequestHeaders:  env.GetList("OPENSTATS_HTTPLOG_REQUEST_HEADERS"),
		LogResponseHeaders: env.GetList("OPENSTATS_HTTPLOG_RESPONSE_HEADERS"),
	}

	if env.GetBool("OPENSTATS_HTTPLOG_REQUEST_BODIES") {
		options.LogRequestBody = func(r *http.Request) bool { return true }
	}

	if env.GetBool("OPENSTATS_HTTPLOG_RESPONSE_BODIES") {
		options.LogResponseBody = func(r *http.Request) bool { return true }
	}

	router.Use(httplog.RequestLogger(logger, options))
	router.Use(cors.New(cors.Options{
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowCredentials: true,
	}).Handler)

	// TODO: CSRF middleware
	// TODO: rate limit middleware

	return router, nil
}

func main() {
	if err := env.Load(); err != nil {
		golog.Fatalf("Error loading envvars from .env files: %v", err)
	}

	env.Require(
		"OPENSTATS_DB_HOST",
		"OPENSTATS_DB_PORT",
		"OPENSTATS_DB_NAME",
		"OPENSTATS_DB_USERNAME",
		"OPENSTATS_DB_PASSWORD",
		"OPENSTATS_DB_TRACE_LOG",
		"OPENSTATS_SLOG_LEVEL",
		"OPENSTATS_SLOG_MODE",
		"OPENSTATS_APP_BASEURL",
		"OPENSTATS_HTTP_ADDR",
		"OPENSTATS_HTTPLOG_LEVEL",
		"OPENSTATS_HTTPLOG_MODE",
		"OPENSTATS_HTTPLOG_CONCISE",
		"OPENSTATS_HTTPLOG_REQUEST_HEADERS",
		"OPENSTATS_HTTPLOG_RESPONSE_HEADERS",
		"OPENSTATS_HTTPLOG_REQUEST_BODIES",
		"OPENSTATS_HTTPLOG_RESPONSE_BODIES",
	)

	if err := log.Setup(); err != nil {
		golog.Fatal(err)
	}

	if err := mail.Setup(context.Background()); err != nil {
		golog.Fatal(err)
	}

	if err := db.SetupDB(context.Background()); err != nil {
		golog.Fatal(err)
	}

	// TODO: we probably aren't using this anymore, after switching to huma...
	if err := validation.SetupValidations(); err != nil {
		golog.Fatal(err)
	}

	// we need a root admin user in order to do admin operations. The root user is also the only user that can add
	// other admins
	if err := auth.AddRootAdminUser(context.Background()); err != nil {
		golog.Fatal(err)
	}

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
		"SessionCookie": {
			Type:        "apiKey",
			In:          "cookie",
			Description: "A session cookie is used to store the user's session token. See Authentication for usage.",
			Name:        auth.SessionCookieName,
		},
		"GameToken": {
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "RID",
			Description:  "A secret token generated by a user, to authenticate a game to track game stats, sessions, and achievements for them.",
		},
		"GameSession": {
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "JWT",
			Description:  "A JWT created when a Game Session is started, used to authenticate all Game Session actions.",
		},
	}

	router, routerErr := setupRouter()
	if routerErr != nil {
		golog.Fatal(routerErr)
	}

	api := humachi.New(router, config)

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

	// TODO: setup remote media when not local
	media.SetupLocal(api)
	users.RegisterRoutes(api)
	internal.RegisterRoutes(api)

	address := env.GetString("OPENSTATS_HTTP_ADDR")
	if err := http.ListenAndServe(address, router); err != nil {
		golog.Fatal(err)
	}
}

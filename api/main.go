package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/dresswithpockets/openstats/app/password"
	"github.com/dresswithpockets/openstats/app/problems"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/template/jet/v2"
)

const (
	MaxDisplayNameLength = 64
	MinDisplayNameLength = 1
	MaxSlugNameLength    = 64
	MinSlugNameLength    = 2
	MaxPasswordLength    = 32
	MinPasswordLength    = 10
)

var ValidSlugSpecialCharacters = []rune("!@#$%^&*")

var ArgonParameters = password.Parameters{
	Iterations:  2,
	Memory:      19 * 1024,
	Parallelism: 1,
	SaltLength:  16,
	KeyLength:   32,
}

type RegisterDto struct {
	// Email is optional, and just used for resetting the user's password
	Email string `json:"email,omitempty"`

	// DisplayName is optional, and is only used when displaying their profile on the website
	DisplayName string `json:"displayName,omitempty"`

	// Slug is a unique username for the user
	Slug string `json:"slug"`

	// Password is the user's login password
	Password string `json:"password"`
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

	templateEngine := jet.New("./views", ".jet.html")

	templateEngine.Reload(true)

	app := fiber.New(fiber.Config{
		Views: templateEngine,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			var validationErr *ValidationError
			if errors.As(err, &validationErr) {
				var fieldErrors map[string][]string
				for _, fieldError := range validationErr.Errors {
					detail := GetValidationDetail(fieldError.Field)
					fieldErrors[fieldError.Field] = append(fieldErrors[fieldError.Field], detail)
				}

				c.Status(400)
				return c.JSON(problems.Validation("", fieldErrors))
			}

			var conflictErr *ConflictError
			if errors.As(err, &conflictErr) {
				c.Status(fiber.StatusConflict)
				return c.JSON(problems.Conflict(conflictErr.Field, conflictErr.Value, ""))
			}

			return fiber.DefaultErrorHandler(c, err)
		},
	})

	app.Use(cors.New())

	// TODO: csrf in local ?
	//app.Use(csrf.New())

	app.Use(limiter.New(limiter.Config{
		Max:               30,
		Expiration:        30 * time.Second,
		LimiterMiddleware: limiter.SlidingWindow{},
	}))

	app.Use(healthcheck.New())

	app.Use(favicon.New(favicon.Config{File: "static/favicon.ico"}))

	app.Route("static", func(router fiber.Router) {
		router.Use(filesystem.New(filesystem.Config{
			Root: http.Dir("./static"),
		}))
	})

	app.Use("/", AuthHandler)
	app.Get("/", viewHomeGet)

	if err := SetupAuthRoutes(app); err != nil {
		log.Fatal(err)
	}

	if err := SetupAdminViews(app); err != nil {
		log.Fatal(err)
	}

	if err := SetupUsersRoutes(app); err != nil {
		log.Fatal(err)
	}

	// TODO: user or profile apis
	// TODO: developer apis
	// TODO: developer-game apis
	// TODO: developer-game-achievement apis

	log.Fatal(app.Listen(":3000"))
}

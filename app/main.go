package main

import (
	"github.com/dresswithpockets/openstats/app/password"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/template/jet/v2"
	"log"
	"time"
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
	Email *string `json:"email,omitempty"`

	// DisplayName is optional, and is only used when displaying their profile on the website
	DisplayName *string `json:"displayName,omitempty"`

	// Slug is a unique username for the user
	Slug string `json:"slug"`

	// Password is the user's login password
	Password string `json:"password"`
}

func main() {
	if err := SetupDB(); err != nil {
		log.Fatal(err)
	}

	if err := SetupAuth(); err != nil {
		log.Fatal(err)
	}

	templateEngine := jet.New("./views", ".jet.html")

	app := fiber.New(fiber.Config{
		Views: templateEngine,
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
	app.Use(favicon.New())
	app.Use("/", AuthHandler)
	app.Get("/", func(c *fiber.Ctx) error {
		userSlug := ""
		user, userIsUser := c.Locals("user").(*User)
		if userIsUser {
			userSlug = user.Slug
		}

		return c.Render("index", fiber.Map{
			"HasSession": userIsUser,
			"UserSlug":   userSlug,
		})
	})

	if err := SetupAuthRoutes(app); err != nil {
		log.Fatal(err)
	}

	//if err := SetupAuthApi(app); err != nil {
	//	log.Fatal(err)
	//}
	//
	//if err := SetupUserApi(app); err != nil {
	//	log.Fatal(err)
	//}

	// TODO: user or profile apis
	// TODO: developer apis
	// TODO: developer-game apis
	// TODO: developer-game-achievement apis

	log.Fatal(app.Listen(":3000"))
}

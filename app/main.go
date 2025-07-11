package main

import (
	"errors"
	"github.com/dresswithpockets/openstats/app/password"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/template/jet/v2"
	"gorm.io/gorm"
	"log"
	"slices"
	"time"
	"unicode"
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

type AuthUserResponse struct {
	Slug string
}

func ValidDisplayName(displayName string) bool {
	return len(displayName) >= MinDisplayNameLength && len(displayName) <= MaxDisplayNameLength
}

// ValidSlug returns true if all of these rules are followed:
//   - slug is at least MinSlugNameLength and no more than MaxSlugNameLength in length
//   - slug is all lowercase
//   - slug contains only latin characters, numbers, or a dash
func ValidSlug(slug string) bool {
	if len(slug) < MinSlugNameLength || len(slug) > MaxSlugNameLength {
		return false
	}

	for _, r := range []rune(slug) {
		if !unicode.IsLower(r) && !unicode.IsNumber(r) && !unicode.IsLetter(r) && r != '-' {
			return false
		}
	}

	return true
}

// ValidPassword returns true if all of these rules are followed:
//   - password is at least MinPasswordLength and no more than MaxPasswordLength in length
//   - password contains only latin characters, numbers, or some special characters: !@#$%^&*
func ValidPassword(password string) bool {
	if len(password) < MinPasswordLength || len(password) > MaxPasswordLength {
		return false
	}

	for _, r := range []rune(password) {
		if !unicode.IsNumber(r) && !unicode.IsLetter(r) && !slices.Contains(ValidSlugSpecialCharacters, r) {
			return false
		}
	}

	return true
}

func NewAuth(db *gorm.DB, store *session.Store) fiber.Handler {
	return func(c *fiber.Ctx) error {
		currentSession, err := store.Get(c)
		if err != nil {
			log.Println(err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		sessionUserId := currentSession.Get("userId")
		if sessionUserId == nil {
			return c.Next()
		}

		var sessionUser User
		result := db.First(&sessionUser, sessionUserId)
		if result.Error == nil {
			c.Locals("user", &sessionUser)
		}

		if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			log.Println(result.Error)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		return c.Next()
	}
}

func NewRequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user")
		if user == nil {
			return c.SendStatus(fiber.StatusUnauthorized)
		}

		if _, isUser := user.(User); !isUser {
			return c.SendStatus(fiber.StatusUnauthorized)
		}

		return c.Next()
	}
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

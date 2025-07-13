package main

import (
	"errors"
	"github.com/dresswithpockets/openstats/app/password"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/fiber/v2/utils"
	sqliteStorage "github.com/gofiber/storage/sqlite3/v2"
	"gorm.io/gorm"
	"log"
	"net/mail"
	"slices"
	"time"
	"unicode"
)

var AuthHandler fiber.Handler
var SessionStore *session.Store
var authSetupComplete = false

func NewAuthHandler(db *gorm.DB, store *session.Store) fiber.Handler {
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

func RequireAuthHandler(c *fiber.Ctx) error {
	user := c.Locals("user")
	if user == nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	if _, isUser := user.(*User); !isUser {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	return c.Next()
}

func RequireAdminAuthHandler(c *fiber.Ctx) error {
	userLocal := c.Locals("user")
	if userLocal == nil {
		return c.SendStatus(fiber.StatusNotFound)
	}

	if user, isUser := userLocal.(*User); !isUser || !IsAdmin(user) {
		return c.SendStatus(fiber.StatusNotFound)
	}

	return c.Next()
}

func SetupAuth() error {
	if DB == nil {
		return errors.New("DB not initialized")
	}

	SessionStore = session.New(session.Config{
		Expiration:   7 * 24 * time.Hour,
		Storage:      sqliteStorage.New(sqliteStorage.Config{}),
		CookieSecure: true,
		KeyGenerator: utils.UUIDv4,
	})

	AuthHandler = NewAuthHandler(DB, SessionStore)

	authSetupComplete = true
	return nil
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

func handleGetLoginView(c *fiber.Ctx) error {
	if c.Locals("user") != nil {
		// we're already authorized so we can just go back home
		return c.RedirectBack("/")
	}

	// otherwise, we can just render the login form
	return c.Render("login", nil)
}

func handlePostLoginView(c *fiber.Ctx) error {
	currentSession, err := SessionStore.Get(c)
	if err != nil {
		log.Println(err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	type LoginDto struct {
		// Slug is the user's unique username
		Slug     string `json:"slug" form:"slug"`
		Password string `json:"password" form:"password"`
	}

	var loginBody LoginDto
	if bodyErr := c.BodyParser(&loginBody); bodyErr != nil {
		// TODO: return problem json indicating the error
		// TODO: redirect to `/login` with bad request info
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if !ValidSlug(loginBody.Slug) {
		// TODO: return problem json indicating the error
		// TODO: redirect to `/login` with bad request info
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if !ValidPassword(loginBody.Password) {
		// TODO: return problem json indicating the error
		// TODO: redirect to `/login` with bad request info
		return c.SendStatus(fiber.StatusBadRequest)
	}

	var matchedUser User
	result := DB.Preload("Password").First(&matchedUser, "slug = ?", loginBody.Slug)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// TODO: redirect to `/login` with username not found or password doesnt match
		return c.SendStatus(fiber.StatusNotFound)
	}

	if result.Error != nil {
		log.Println(result.Error)
		return c.Status(fiber.StatusInternalServerError).Render("500", nil)
	}

	verifyErr := password.VerifyPassword(loginBody.Password, matchedUser.Password.EncodedHash)
	if errors.Is(verifyErr, password.ErrHashMismatch) {
		// TODO: redirect to `/login` with username not found or password doesnt match
		return c.SendStatus(fiber.StatusNotFound)
	}

	if verifyErr != nil {
		log.Println(verifyErr)
		return c.Status(fiber.StatusInternalServerError).Render("500", nil)
	}

	currentSession.Set("userId", matchedUser.ID)
	if saveErr := currentSession.Save(); saveErr != nil {
		log.Println(saveErr)
		return c.Status(fiber.StatusInternalServerError).Render("500", nil)
	}

	return c.Redirect("/")
}

func handleGetRegisterView(c *fiber.Ctx) error {
	if c.Locals("user") != nil {
		// we're already authorized so we can just go back home
		return c.Redirect("/")
	}

	// otherwise, we can just render the login form
	return c.Render("register", nil)
}

var (
	ErrInvalidEmailAddress = errors.New("invalid email address")
	ErrInvalidDisplayName  = errors.New("invalid display name")
	ErrInvalidSlug         = errors.New("invalid slug")
	ErrInvalidPassword     = errors.New("invalid password")
	ErrSlugAlreadyInUse    = errors.New("slug already in use")
)

func AddNewUser(displayName, email string, slug, pass string) (newUser *User, err error) {
	if len(email) > 0 {
		_, emailErr := mail.ParseAddress(email)
		if emailErr != nil {
			return nil, ErrInvalidEmailAddress
		}
	}

	if len(displayName) > 0 {
		if !ValidDisplayName(displayName) {
			// TODO: return problem json indicating the error
			// TODO: redirect to `/register` with bad request info
			return nil, ErrInvalidDisplayName
		}
	}

	if !ValidSlug(slug) {
		return nil, ErrInvalidSlug
	}

	if !ValidPassword(pass) {
		return nil, ErrInvalidPassword
	}

	encodedPassword, passwordErr := password.EncodePassword(pass, ArgonParameters)
	if passwordErr != nil {
		log.Println(passwordErr)
		return nil, passwordErr
	}

	newUser = &User{
		Slug:        slug,
		Password:    UserPassword{EncodedHash: encodedPassword},
		SlugRecords: []UserSlugRecord{{Slug: slug}},
	}

	if len(email) > 0 {
		newUser.Emails = []UserEmail{{Email: email}}
	}

	if len(displayName) > 0 {
		newUser.DisplayNames = []UserDisplayName{{Name: displayName}}
	}

	result := DB.Create(newUser)
	if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
		// TODO: redirect to `/register` with conflict info
		return nil, ErrSlugAlreadyInUse
	}

	if result.Error != nil {
		log.Println(result.Error)
		return nil, result.Error
	}

	return newUser, nil
}

func handlePostRegisterView(c *fiber.Ctx) error {
	if c.Locals("user") != nil {
		// we're already authorized so we can just go back home
		return c.Redirect("/")
	}

	currentSession, err := SessionStore.Get(c)
	if err != nil {
		log.Println(err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	var registerBody RegisterDto
	if bodyErr := c.BodyParser(&registerBody); bodyErr != nil {
		// TODO: return problem json indicating the error
		// TODO: redirect to `/register` with bad request info
		return c.SendStatus(fiber.StatusBadRequest)
	}

	newUser, newUserError := AddNewUser(registerBody.DisplayName, registerBody.Email, registerBody.Slug, registerBody.Password)
	if newUserError != nil {
		if errors.Is(newUserError, ErrInvalidEmailAddress) {
			// TODO: return problem json indicating the error
			// TODO: redirect to `/register` with bad request info
			return c.SendStatus(fiber.StatusBadRequest)
		}

		if errors.Is(newUserError, ErrInvalidDisplayName) {
			// TODO: return problem json indicating the error
			// TODO: redirect to `/register` with bad request info
			return c.SendStatus(fiber.StatusBadRequest)
		}

		if errors.Is(newUserError, ErrInvalidSlug) {
			// TODO: return problem json indicating the error
			// TODO: redirect to `/register` with bad request info
			return c.SendStatus(fiber.StatusBadRequest)
		}

		if errors.Is(newUserError, ErrInvalidPassword) {
			// TODO: return problem json indicating the error
			// TODO: redirect to `/register` with bad request info
			return c.SendStatus(fiber.StatusBadRequest)
		}

		if errors.Is(newUserError, ErrSlugAlreadyInUse) {
			return c.SendStatus(fiber.StatusConflict)
		}

		log.Println(newUserError)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	currentSession.Set("userId", newUser.ID)
	if saveErr := currentSession.Save(); saveErr != nil {
		log.Println(saveErr)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.Redirect("/")
}

func handleGetLogoutView(c *fiber.Ctx) error {
	if c.Locals("user") == nil {
		// we're already logged out, just go back home
		return c.Redirect("/")
	}

	// the logout view will send a logout post request on page load
	return c.Render("logout", nil)
}

func handlePostLogoutView(c *fiber.Ctx) error {
	if c.Locals("user") == nil {
		// we're already logged out, just go back home
		return c.Redirect("/")
	}

	currentSession, err := SessionStore.Get(c)
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusInternalServerError).Render("500", nil)
	}

	destroyErr := currentSession.Destroy()
	if destroyErr != nil {
		log.Println(destroyErr)
		return c.Status(fiber.StatusInternalServerError).Render("500", nil)
	}

	return c.Redirect("/")
}

func SetupAuthRoutes(router fiber.Router) error {
	if !authSetupComplete {
		return errors.New("call SetupAuth() before calling SetupAuthRoutes()")
	}

	rootRoute := router.Group("/")
	rootRoute.Use(AuthHandler)
	rootRoute.Use("/auth/logout", RequireAuthHandler)
	rootRoute.Get("/login", handleGetLoginView)
	rootRoute.Post("/login", handlePostLoginView)
	rootRoute.Get("/register", handleGetRegisterView)
	rootRoute.Post("/register", handlePostRegisterView)
	// since GET requests MUST be idempotent on fly.io, the logout request must be AJAX
	rootRoute.Get("/logout", handleGetLogoutView)
	rootRoute.Post("/logout", handlePostLogoutView)

	return nil
}

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
	"time"
)

var AuthHandler fiber.Handler
var RequireAuthHandler fiber.Handler
var SessionStore *session.Store
var authSetupComplete = false

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

	AuthHandler = NewAuth(DB, SessionStore)
	RequireAuthHandler = NewRequireAuth()

	authSetupComplete = true
	return nil
}

func authUser(c *fiber.Ctx) error {
	user, userIsUser := c.Locals("user").(*User)
	if !userIsUser {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	return c.JSON(AuthUserResponse{
		Slug: user.Slug,
	}, "application/json")
}

func authRegister(c *fiber.Ctx) error {
	c.Accepts("application/json")

	if c.Locals("user") != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	currentSession, err := SessionStore.Get(c)
	if err != nil {
		log.Println(err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	var registerBody RegisterDto
	if bodyErr := c.BodyParser(&registerBody); bodyErr != nil {
		// TODO: return problem json indicating the error
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if registerBody.Email != nil {
		_, emailErr := mail.ParseAddress(*registerBody.Email)
		if emailErr != nil {
			// TODO: return problem json indicating the error
			return c.SendStatus(fiber.StatusBadRequest)
		}
	}

	if registerBody.DisplayName != nil {
		if !ValidDisplayName(*registerBody.DisplayName) {
			// TODO: return problem json indicating the error
			return c.SendStatus(fiber.StatusBadRequest)
		}
	}

	if !ValidSlug(registerBody.Slug) {
		// TODO: return problem json indicating the error
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if !ValidPassword(registerBody.Password) {
		// TODO: return problem json indicating the error
		return c.SendStatus(fiber.StatusBadRequest)
	}

	encodedPassword, passwordErr := password.EncodePassword(registerBody.Password, ArgonParameters)
	if passwordErr != nil {
		log.Println(passwordErr)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	newUser := User{
		Slug:        registerBody.Slug,
		Password:    UserPassword{EncodedHash: encodedPassword},
		SlugRecords: []UserSlugRecord{{Slug: registerBody.Slug}},
	}

	if registerBody.Email != nil {
		newUser.Emails = []UserEmail{{Email: *registerBody.Email}}
	}

	if registerBody.DisplayName != nil {
		newUser.DisplayNames = []UserDisplayName{{Name: *registerBody.DisplayName}}
	}

	result := DB.Create(&newUser)
	if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
		return c.SendStatus(fiber.StatusConflict)
	}

	if result.Error != nil {
		log.Println(result.Error)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	currentSession.Set("userId", newUser.ID)
	if saveErr := currentSession.Save(); saveErr != nil {
		log.Println(saveErr)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(AuthUserResponse{
		Slug: newUser.Slug,
	}, "application/json")
}

func authLogin(c *fiber.Ctx) error {
	if c.Locals("user") != nil {
		// we're already authorized!!!
		return c.Redirect("/")
	}

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
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if !ValidSlug(loginBody.Slug) {
		// TODO: return problem json indicating the error
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if !ValidPassword(loginBody.Password) {
		// TODO: return problem json indicating the error
		return c.SendStatus(fiber.StatusBadRequest)
	}

	var matchedUser User
	result := DB.Preload("Password").First(&matchedUser, "slug = ?", loginBody.Slug)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return c.SendStatus(fiber.StatusNotFound)
	}

	if result.Error != nil {
		log.Println(result.Error)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	verifyErr := password.VerifyPassword(loginBody.Password, matchedUser.Password.EncodedHash)
	if errors.Is(verifyErr, password.ErrHashMismatch) {
		return c.SendStatus(fiber.StatusNotFound)
	}

	if verifyErr != nil {
		log.Println(verifyErr)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	currentSession.Set("userId", matchedUser.ID)
	if saveErr := currentSession.Save(); saveErr != nil {
		log.Println(saveErr)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.Redirect("/")
}

func authLogout(c *fiber.Ctx) error {
	currentSession, err := SessionStore.Get(c)
	if err != nil {
		log.Println(err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	destroyErr := currentSession.Destroy()
	if destroyErr != nil {
		log.Println(destroyErr)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.Redirect("/")
}

func authCreateToken(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNotFound)
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

	if registerBody.Email != nil && len(*registerBody.Email) > 0 {
		_, emailErr := mail.ParseAddress(*registerBody.Email)
		if emailErr != nil {
			// TODO: return problem json indicating the error
			// TODO: redirect to `/register` with bad request info
			return c.SendStatus(fiber.StatusBadRequest)
		}
	}

	if registerBody.DisplayName != nil && len(*registerBody.DisplayName) > 0 {
		if !ValidDisplayName(*registerBody.DisplayName) {
			// TODO: return problem json indicating the error
			// TODO: redirect to `/register` with bad request info
			return c.SendStatus(fiber.StatusBadRequest)
		}
	}

	if !ValidSlug(registerBody.Slug) {
		// TODO: return problem json indicating the error
		// TODO: redirect to `/register` with bad request info
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if !ValidPassword(registerBody.Password) {
		// TODO: return problem json indicating the error
		// TODO: redirect to `/register` with bad request info
		return c.SendStatus(fiber.StatusBadRequest)
	}

	encodedPassword, passwordErr := password.EncodePassword(registerBody.Password, ArgonParameters)
	if passwordErr != nil {
		log.Println(passwordErr)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	newUser := User{
		Slug:        registerBody.Slug,
		Password:    UserPassword{EncodedHash: encodedPassword},
		SlugRecords: []UserSlugRecord{{Slug: registerBody.Slug}},
	}

	if registerBody.Email != nil {
		newUser.Emails = []UserEmail{{Email: *registerBody.Email}}
	}

	if registerBody.DisplayName != nil {
		newUser.DisplayNames = []UserDisplayName{{Name: *registerBody.DisplayName}}
	}

	result := DB.Create(&newUser)
	if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
		// TODO: redirect to `/register` with conflict info
		return c.SendStatus(fiber.StatusConflict)
	}

	if result.Error != nil {
		log.Println(result.Error)
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

func SetupAuthApi(app fiber.Router) error {
	if !authSetupComplete {
		return errors.New("call SetupAuth() before calling SetupAuthApi()")
	}

	apiGroup := app.Group("/api")
	apiGroup.Use(AuthHandler)
	apiGroup.Use([]string{"/auth/user", "/auth/logout"}, RequireAuthHandler)
	apiGroup.Get("/auth/user", authUser)
	apiGroup.Post("/auth/register", authRegister)
	apiGroup.Post("/auth/login", authLogin)
	apiGroup.Post("/auth/logout", authLogout)
	apiGroup.Post("/auth/create-token", authCreateToken)

	return nil
}

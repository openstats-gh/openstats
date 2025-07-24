package main

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/dresswithpockets/openstats/app/password"
	"github.com/dresswithpockets/openstats/app/queries"
	"github.com/dresswithpockets/openstats/app/query"
	"github.com/gofiber/fiber/v2"
)

func AuthHandler(c *fiber.Ctx) error {
	currentSession, err := SessionStore.Get(c)
	if err != nil {
		log.Println(err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	sessionUserId, ok := currentSession.GetUserID()
	if !ok {
		return c.Next()
	}

	sessionUser, findErr := Queries.FindUser(c.Context(), sessionUserId)
	if errors.Is(findErr, sql.ErrNoRows) {
		return c.Next()
	}

	if findErr != nil {
		log.Println(findErr)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	Locals.User.Set(c, &sessionUser)
	return c.Next()
}

func RequireAuthHandler(c *fiber.Ctx) error {
	if !Locals.User.Exists(c) {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	return c.Next()
}

func RequireAdminAuthHandler(c *fiber.Ctx) error {
	user, ok := Locals.User.Get(c)
	if !ok || !IsAdmin(user) {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	return c.Next()
}

func handleGetLoginView(c *fiber.Ctx) error {
	if Locals.User.Exists(c) {
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
		Slug     string `json:"slug" form:"slug" validator:"required,slug"`
		Password string `json:"password" form:"password" validator:"required,password"`
	}

	var loginBody LoginDto
	if bodyErr := c.BodyParser(&loginBody); bodyErr != nil {
		// TODO: return problem json indicating the error
		// TODO: redirect to `/login` with bad request info
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if validateErr := Validate(loginBody); validateErr != nil {
		return validateErr
	}

	result, findErr := Queries.FindUserBySlugWithPassword(c.Context(), loginBody.Slug)
	if errors.Is(findErr, sql.ErrNoRows) {
		// TODO: redirect to `/login` with username not found or password doesnt match
		return c.SendStatus(fiber.StatusNotFound)
	}

	if findErr != nil {
		log.Println(findErr)
		return c.Status(fiber.StatusInternalServerError).Render("500", nil)
	}

	verifyErr := password.VerifyPassword(loginBody.Password, result.EncodedHash)
	if errors.Is(verifyErr, password.ErrHashMismatch) {
		// TODO: redirect to `/login` with username not found or password doesnt match
		return c.SendStatus(fiber.StatusNotFound)
	}

	if verifyErr != nil {
		log.Println(verifyErr)
		return c.Status(fiber.StatusInternalServerError).Render("500", nil)
	}

	currentSession.SetUserID(result.ID)
	if saveErr := currentSession.Save(); saveErr != nil {
		log.Println(saveErr)
		return c.Status(fiber.StatusInternalServerError).Render("500", nil)
	}

	return c.Redirect("/")
}

func handleGetRegisterView(c *fiber.Ctx) error {
	if Locals.User.Exists(c) {
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
)

func AddNewUser(ctx context.Context, displayName, email string, slug, pass string) (newUser *query.User, err error) {
	if len(email) > 0 && !ValidEmailAddress(email) {
		return nil, ErrInvalidEmailAddress
	}

	if len(displayName) > 0 && !ValidDisplayName(displayName) {
		return nil, ErrInvalidDisplayName
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

	return Actions.CreateUser(ctx, slug, encodedPassword, email, displayName)
}

func handlePostRegisterView(c *fiber.Ctx) error {
	if Locals.User.Exists(c) {
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

	newUser, newUserError := AddNewUser(
		c.Context(),
		registerBody.DisplayName,
		registerBody.Email,
		registerBody.Slug,
		registerBody.Password,
	)
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

		if errors.Is(newUserError, queries.ErrSlugAlreadyInUse) {
			return c.SendStatus(fiber.StatusConflict)
		}

		log.Println(newUserError)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	currentSession.SetUserID(newUser.ID)
	if saveErr := currentSession.Save(); saveErr != nil {
		log.Println(saveErr)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.Redirect("/")
}

func handleGetLogoutView(c *fiber.Ctx) error {
	if !Locals.User.Exists(c) {
		// we're already logged out, just go back home
		return c.Redirect("/")
	}

	// the logout view will send a logout post request on page load
	return c.Render("logout", nil)
}

func handlePostLogoutView(c *fiber.Ctx) error {
	if !Locals.User.Exists(c) {
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

func handleGetSession(c *fiber.Ctx) error {
	localUser, userExists := Locals.User.Get(c)
	if !userExists {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	userDisplayName, err := Queries.GetUserLatestDisplayName(c.Context(), localUser.ID)
	if errors.Is(err, sql.ErrNoRows) {
		userDisplayName = query.UserDisplayName{DisplayName: ""}
	} else if err != nil {
		log.Println(err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	// otherwise, we can just render the login form
	return c.JSON(fiber.Map{
		"slug":        localUser.Slug,
		"displayName": userDisplayName.DisplayName,
	})
}

func SetupAuthRoutes(router fiber.Router) error {
	rootRoute := router.Group("/auth")

	rootRoute.Use(AuthHandler)
	rootRoute.Use("/sign-out", RequireAuthHandler)
	rootRoute.Get("/session", handleGetSession)
	rootRoute.Post("/sign-in", handlePostLoginView)
	rootRoute.Post("/sign-up", handlePostRegisterView)
	rootRoute.Post("/sign-out", handlePostLogoutView)

	// rootRoute.Use("/auth/logout", RequireAuthHandler)
	// rootRoute.Get("/login", handleGetLoginView)
	// rootRoute.Post("/login", handlePostLoginView)
	// rootRoute.Get("/register", handleGetRegisterView)
	// rootRoute.Post("/register", handlePostRegisterView)
	// since GET requests MUST be idempotent on fly.io, the logout request must be AJAX
	// rootRoute.Get("/logout", handleGetLogoutView)
	// rootRoute.Post("/logout", handlePostLogoutView)

	return nil
}

package main

import (
	"context"
	"database/sql"
	"errors"
	"github.com/dresswithpockets/openstats/app/db"
	"github.com/dresswithpockets/openstats/app/db/query"
	"github.com/dresswithpockets/openstats/app/password"
	"github.com/gofiber/fiber/v2"
	"github.com/rotisserie/eris"
)

func AuthHandler(c *fiber.Ctx) error {
	currentSession, err := SessionStore.Get(c)
	if err != nil {
		return err
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
		return findErr
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

func HandlePostSignIn(c *fiber.Ctx) error {
	currentSession, err := SessionStore.Get(c)
	if err != nil {
		return err
	}

	type LoginDto struct {
		// Slug is the user's unique username
		Slug     string `json:"slug" validator:"required,slug"`
		Password string `json:"password" validator:"required,password"`
	}

	var loginBody LoginDto
	if bodyErr := c.BodyParser(&loginBody); bodyErr != nil {
		// the body is either invalid json or otherwise unparsable with LoginDto
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if validateErr := Validate(loginBody); validateErr != nil {
		return validateErr
	}

	result, findErr := Queries.FindUserBySlugWithPassword(c.Context(), loginBody.Slug)
	if errors.Is(findErr, sql.ErrNoRows) {
		return c.SendStatus(fiber.StatusNotFound)
	}

	if findErr != nil {
		return findErr
	}

	verifyErr := password.VerifyPassword(loginBody.Password, result.EncodedHash)
	if errors.Is(verifyErr, password.ErrHashMismatch) {
		return c.SendStatus(fiber.StatusNotFound)
	}

	if verifyErr != nil {
		return verifyErr
	}

	currentSession.SetUserID(result.ID)
	if saveErr := currentSession.Save(); saveErr != nil {
		return saveErr
	}

	return nil
}

var (
	ErrInvalidEmailAddress = errors.New("invalid email address")
	ErrInvalidDisplayName  = errors.New("invalid display name")
	ErrInvalidSlug         = errors.New("invalid slug")
	ErrInvalidPassword     = errors.New("invalid password")
)

func AddNewUser(ctx context.Context, displayName, email, slug, pass string) (newUser *query.User, err error) {
	if len(email) > 0 && !ValidEmailAddress(email) {
		return nil, eris.Wrap(ErrInvalidEmailAddress, "validation error")
	}

	if len(displayName) > 0 && !ValidDisplayName(displayName) {
		return nil, eris.Wrap(ErrInvalidDisplayName, "validation error")
	}

	if !ValidSlug(slug) {
		return nil, eris.Wrap(ErrInvalidSlug, "validation error")
	}

	if !ValidPassword(pass) {
		return nil, eris.Wrap(ErrInvalidPassword, "validation error")
	}

	encodedPassword, passwordErr := password.EncodePassword(pass, ArgonParameters)
	if passwordErr != nil {
		return nil, passwordErr
	}

	return Actions.CreateUser(ctx, slug, encodedPassword, email, displayName)
}

func HandlePostSignUp(c *fiber.Ctx) error {
	if Locals.User.Exists(c) {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	currentSession, err := SessionStore.Get(c)
	if err != nil {
		return err
	}

	type RegisterDto struct {
		// Email is optional, and just used for resetting the user's password
		Email string `json:"email,omitempty" validator:"email"`

		// DisplayName is optional, and is only used when displaying their profile on the website
		DisplayName string `json:"displayName,omitempty" validator:"displayName"`

		// Slug is a unique username for the user
		Slug string `json:"slug" validator:"required,slug"`

		// Password is the user's login password
		Password string `json:"password" validator:"required,password"`
	}

	var registerBody RegisterDto
	if bodyErr := c.BodyParser(&registerBody); bodyErr != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if validateErr := Validate(registerBody); validateErr != nil {
		return validateErr
	}

	newUser, newUserError := AddNewUser(
		c.Context(),
		registerBody.DisplayName,
		registerBody.Email,
		registerBody.Slug,
		registerBody.Password,
	)
	if newUserError != nil {
		if errors.Is(newUserError, db.ErrSlugAlreadyInUse) {
			return Conflict("slug", registerBody.Slug)
		}

		if ErrorIsAny(newUserError, ErrInvalidEmailAddress, ErrInvalidDisplayName, ErrInvalidSlug, ErrInvalidPassword) {
			return eris.Wrap(newUserError, "registerBody was Validated, and yet AddUser returned a validation error")
		}

		return newUserError
	}

	currentSession.SetUserID(newUser.ID)
	if saveErr := currentSession.Save(); saveErr != nil {
		return saveErr
	}

	return nil
}

func HandlePostSignOut(c *fiber.Ctx) error {
	if !Locals.User.Exists(c) {
		// we're already logged out, just go back home
		return nil
	}

	currentSession, err := SessionStore.Get(c)
	if err != nil {
		return err
	}

	destroyErr := currentSession.Destroy()
	if destroyErr != nil {
		return destroyErr
	}

	return nil
}

func HandleGetSession(c *fiber.Ctx) error {
	localUser, userExists := Locals.User.Get(c)
	if !userExists {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	userDisplayName, err := Queries.GetUserLatestDisplayName(c.Context(), localUser.ID)
	if errors.Is(err, sql.ErrNoRows) {
		userDisplayName = query.UserDisplayName{DisplayName: ""}
	} else if err != nil {
		return err
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
	rootRoute.Get("/session", HandleGetSession)
	rootRoute.Post("/sign-in", HandlePostSignIn)
	rootRoute.Post("/sign-up", HandlePostSignUp)
	rootRoute.Use("/sign-out", RequireAuthHandler)
	rootRoute.Post("/sign-out", HandlePostSignOut)

	return nil
}

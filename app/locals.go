package main

import (
	"errors"
	"github.com/dresswithpockets/openstats/app/models"
	"github.com/gofiber/fiber/v2"
)

var (
	ErrLocalNotFound       = errors.New("local not found")
	ErrLocalNotCorrectType = errors.New("local not correct type")
)

func LocalSetUser(c *fiber.Ctx, user *models.User) {
	c.Locals("user", user)
}

func LocalGetUser(c *fiber.Ctx) (*models.User, error) {
	localUser := c.Locals("user")
	if localUser == nil {
		return nil, ErrLocalNotFound
	}

	asUser, asUserOk := localUser.(*models.User)
	if !asUserOk {
		return nil, ErrLocalNotCorrectType
	}

	return asUser, nil
}

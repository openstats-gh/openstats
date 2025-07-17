package main

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"log"
)

type GetUserResponse struct {
	Slug        string  `json:"slug"`
	DisplayName *string `json:"display_name,omitempty"`
}

func userGetUser(c *fiber.Ctx) error {
	slug := c.Params("slug")
	if slug == "" {
		return c.SendStatus(fiber.StatusNotFound)
	}

	if !ValidSlug(slug) {
		return c.SendStatus(fiber.StatusNotFound)
	}

	var response GetUserResponse
	result := GormDB.Model(&User{}).
		Select("users.slug, udn.name as display_name").
		Joins("left outer joins user_display_names udn on users.id = udn.user_id").
		Where(&User{Slug: slug}).
		Order("udn.name desc").
		Limit(1).
		Scan(&response)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return c.SendStatus(fiber.StatusNotFound)
	}

	if result.Error != nil {
		log.Println(result.Error)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(response)
}

func userGetAchievements(c *fiber.Ctx) error {
	return nil
}

func viewUserGet(c *fiber.Ctx) error {
	return nil
}

func SetupUserViews(router fiber.Router) {
	userGroup := router.Group("/user")
	userGroup.Get("/:userSlug", viewUserGet)
}

func SetupUserApi(router fiber.Router) error {
	userGroup := router.Group("/user")
	userGroup.Get("/:userSlug", userGetUser)
	userGroup.Get("/:userSlug/achievements", userGetAchievements)
	return nil
}

package main

import (
	"database/sql"
	"errors"
	"github.com/dresswithpockets/openstats/app/db/query"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

func handleGetUsersBrief(c *fiber.Ctx) error {
	userSlug := c.Params("slug")
	if userSlug == "" {
		return c.SendStatus(fiber.StatusNotFound)
	}

	recentUserAchievements, err := Queries.GetUserRecentAchievements(c.Context(), query.GetUserRecentAchievementsParams{
		UserSlug: userSlug,
		Limit:    20,
	})

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		// TODO: show error in the view
		log.Error(err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	recentOtherUserAchievements, err := Queries.GetOtherUserRecentAchievements(c.Context(), query.GetOtherUserRecentAchievementsParams{
		ExcludedUserSlug: userSlug,
		Limit:            20,
	})

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		// TODO: show error in the view
		log.Error(err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	type UnlockedAchievementInfo struct {
		DeveloperSlug string `json:"developerSlug"`
		GameSlug      string `json:"gameSlug"`
		GameName      string `json:"gameName"`
		Slug          string `json:"slug"`
		Name          string `json:"name"`
		Description   string `json:"description"`
	}

	type OtherUserUnlockedAchievementInfo struct {
		DeveloperSlug   string `json:"developerSlug"`
		GameSlug        string `json:"gameSlug"`
		GameName        string `json:"gameName"`
		Slug            string `json:"slug"`
		Name            string `json:"name"`
		Description     string `json:"description"`
		UserSlug        string `json:"userSlug,omitempty"`
		UserDisplayName string `json:"userDisplayName,omitempty"`
	}

	type Response struct {
		Unlocks          []UnlockedAchievementInfo          `json:"unlocks"`
		OtherUserUnlocks []OtherUserUnlockedAchievementInfo `json:"otherUserUnlocks"`
	}

	var unlocks []UnlockedAchievementInfo
	for _, row := range recentUserAchievements {
		unlocks = append(unlocks, UnlockedAchievementInfo{
			DeveloperSlug: row.DeveloperSlug,
			GameSlug:      row.GameSlug,
			GameName:      row.GameName,
			Slug:          userSlug,
			Name:          row.Name,
			Description:   row.Description,
		})
	}

	var otherUserUnlocks []OtherUserUnlockedAchievementInfo
	for _, row := range recentOtherUserAchievements {
		userDisplayName := ""
		if row.UserDisplayName.Valid {
			userDisplayName = row.UserDisplayName.String
		}

		otherUserUnlocks = append(otherUserUnlocks, OtherUserUnlockedAchievementInfo{
			DeveloperSlug:   row.DeveloperSlug,
			GameSlug:        row.GameSlug,
			GameName:        row.GameName,
			Slug:            userSlug,
			Name:            row.Name,
			Description:     row.Description,
			UserSlug:        row.UserSlug,
			UserDisplayName: userDisplayName,
		})
	}

	return c.JSON(Response{
		Unlocks:          unlocks,
		OtherUserUnlocks: otherUserUnlocks,
	})
}

func SetupUsersRoutes(router fiber.Router) error {
	usersRoutes := router.Group("/users")

	usersRoutes.Use(AuthHandler)
	usersRoutes.Get("/:slug/brief", handleGetUsersBrief)

	return nil
}

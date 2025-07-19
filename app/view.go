package main

import (
	"github.com/dresswithpockets/openstats/app/query"
	"github.com/gofiber/fiber/v2"
)

func viewHomeGet(c *fiber.Ctx) error {
	user, hasUser := Locals.User.Get(c)
	userSlug := ""
	var recentUserAchievements []query.GetUserRecentAchievementsRow
	var recentOtherUserAchievements []query.GetOtherUserRecentAchievementsRow
	if hasUser {
		userSlug = user.Slug

		var err error
		recentUserAchievements, err = Queries.GetUserRecentAchievements(c.Context(), query.GetUserRecentAchievementsParams{
			UserID: user.ID,
			Limit:  20,
		})

		if err != nil {
			// TODO: show error in the view
		}

		recentOtherUserAchievements, err = Queries.GetOtherUserRecentAchievements(c.Context(), query.GetOtherUserRecentAchievementsParams{
			UserID: user.ID,
			Limit:  20,
		})

		if err != nil {
			// TODO: show error in the view
		}
	}

	return c.Render("index", fiber.Map{
		"Title":                       "openstats",
		"CurrentPath":                 "home",
		"HasSession":                  hasUser,
		"UserSlug":                    userSlug,
		"RecentUserAchievements":      recentUserAchievements,
		"RecentOtherUserAchievements": recentOtherUserAchievements,
	}, "layouts/main")
}

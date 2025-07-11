package main

import "github.com/gofiber/fiber/v2"

func viewHomeGet(c *fiber.Ctx) error {
	type RecentUserAchievement struct {
		Name     string
		GameName string
	}

	type RecentOtherUserAchievement struct {
		UserName string
		Name     string
		GameName string
	}

	userSlug := ""
	user, userIsUser := c.Locals("user").(*User)
	var recentUserAchievements []RecentUserAchievement
	var recentOtherUserAchievements []RecentOtherUserAchievement
	if userIsUser {
		userSlug = user.Slug

		result := DB.Table("achievement_progresses as ap").
			Select("a.name, g.slug as game_name").
			Joins("join achievements a on ap.achievement_id = a.id").
			Joins("join users u on ap.user_id = u.id").
			Joins("join games g on a.game_id = g.id").
			Where("u.id = ? and ap.progress = a.progress_requirement", user.ID).
			Order("ap.created_at desc").
			Limit(20).
			Find(&recentUserAchievements)

		if result.Error != nil {
			// TODO: show error in the view
		}

		result = DB.Table("achievement_progresses as ap").
			Select("u.slug as user_name, a.name, g.slug as game_name").
			Joins("join achievements a on ap.achievement_id = a.id").
			Joins("join users u on ap.user_id = u.id").
			Joins("join games g on a.game_id = g.id").
			Where("u.id != ? and ap.progress = a.progress_requirement", user.ID).
			Order("ap.created_at desc").
			Limit(20).
			Find(&recentOtherUserAchievements)

		if result.Error != nil {
			// TODO: show error in the view
		}

		/*
			result := DB.Model(&User{}).
			Select("users.slug, udn.name as display_name").
			Joins("left outer joins user_display_names udn on users.id = udn.user_id").
			Where(&User{Slug: slug}).
			Order("udn.name desc").
			Limit(1).
			Scan(&response)
		*/
	}

	return c.Render("index", fiber.Map{
		"Title":                       "openstats",
		"CurrentPath":                 "home",
		"HasSession":                  userIsUser,
		"UserSlug":                    userSlug,
		"RecentUserAchievements":      recentUserAchievements,
		"RecentOtherUserAchievements": recentOtherUserAchievements,
	}, "layouts/main")
}

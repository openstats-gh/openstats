package main

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"gorm.io/gorm"
)

const (
	RootUserDisplayName = "Admin"
	RootUserEmail       = ""
	RootUserSlug        = "openstats"
	RootUserPass        = "openstatsadmin"
)

func AddRootAdminUser() {
	_, newUserErr := AddNewUser(RootUserDisplayName, RootUserEmail, RootUserSlug, RootUserPass)
	// this function is expected to be idempotent - if called multiple times, it shouldn't fail even if the admin
	// already exists
	if newUserErr != nil && !errors.Is(newUserErr, ErrSlugAlreadyInUse) {
		log.Fatal(newUserErr)
	}
}

func IsAdmin(user *User) bool {
	// TODO: add distinction between Admin and Root - Root should be able to add non-root Admin users, which can do
	//       everything except add other Admins
	return user != nil && user.Slug == RootUserSlug
}

func IsRoot(user *User) bool {
	return user != nil && user.Slug == RootUserSlug
}

type AdminPathGroup int

const (
	HomeAdminPathGroup AdminPathGroup = iota
	UsersAdminPathGroup
	DevelopersAdminPathGroup
	GamesAdminPathGroup
)

func getAdminPaths(group AdminPathGroup) []fiber.Map {
	return []fiber.Map{
		{
			"IsCurrent": group == HomeAdminPathGroup,
			"Path":      "/admin",
			"Name":      "Home",
		},
		{
			"IsCurrent": group == UsersAdminPathGroup,
			"Path":      "/admin/users",
			"Name":      "Users",
		},
		{
			"IsCurrent": group == DevelopersAdminPathGroup,
			"Path":      "/admin/developers",
			"Name":      "Developers",
		},
		{
			"IsCurrent": group == GamesAdminPathGroup,
			"Path":      "/admin/developers/@/games",
			"Name":      "Games",
		},
	}
}

func viewAdminHomeGet(ctx *fiber.Ctx) error {
	return ctx.Render("admin/home", fiber.Map{
		"Title":    "Admin Dashboard",
		"NavPages": getAdminPaths(HomeAdminPathGroup),
	}, "layouts/admin")
}

func viewAdminUsersList(ctx *fiber.Ctx) error {
	type UsersList struct {
		Name string
		Slug string
	}

	var usersList []UsersList

	// gets every user's slug and their most recent display name
	result := DB.Table("users as u").
		Select("u.slug, udn1.name").
		Joins("join user_display_names udn1 on u.id = udn1.user_id and udn1.deleted_at is null").
		Joins("left outer join user_display_names udn2 on u.id = udn2.user_id and udn2.deleted_at is null and (udn1.created_at < udn2.created_at OR (udn1.created_at = udn2.created_at and udn1.id < udn2.id))").
		Where("udn2.id is null").
		Find(&usersList)

	if result.Error != nil {
		// TODO: handle error
	}

	return ctx.Render("admin/users", fiber.Map{
		"Title":    "Users",
		"NavPages": getAdminPaths(UsersAdminPathGroup),
		"Users":    usersList,
	}, "layouts/admin")
}

func viewAdminUsersRead(ctx *fiber.Ctx) error {
	slug := ctx.Params("slug")
	if slug == "" || slug == "@" {
		return ctx.Redirect("/admin/users")
	}

	var queriedUser User
	result := DB.Unscoped().
		Model(&User{}).
		Preload("DisplayNames").
		Where(&User{Slug: slug}).
		First(&queriedUser)

	//if errors.Is(result.Error, gorm.ErrRecordNotFound) {
	//	found = false
	//	ctx.Status(fiber.StatusNotFound)
	//	return ctx.Render("admin/user", fiber.Map{
	//		"Title":    "User",
	//		"NavPages": getAdminPaths(UsersAdminPathGroup),
	//		"Found":    false,
	//		"User":     queriedUser,
	//		"PathSlug": slug,
	//	}, "layouts/admin")
	//}

	foundUser := true

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			ctx.Status(fiber.StatusNotFound)
			foundUser = false
		} else {
			// TODO: handle error
		}
	}

	return ctx.Render("admin/user", fiber.Map{
		"Title":    "User",
		"NavPages": getAdminPaths(UsersAdminPathGroup),
		"Found":    foundUser,
		"User":     queriedUser,
		"PathSlug": slug,
	}, "layouts/admin")
}

func viewAdminUsersCreateOrUpdate(ctx *fiber.Ctx) error {
	return nil
}

func viewAdminUsersDelete(ctx *fiber.Ctx) error {
	return nil
}

func viewAdminDevelopersList(ctx *fiber.Ctx) error {
	return ctx.Render("admin/developers", fiber.Map{
		"Title":    "Developers",
		"NavPages": getAdminPaths(DevelopersAdminPathGroup),
	}, "layouts/admin")
}

func viewAdminDevelopersRead(ctx *fiber.Ctx) error {
	return nil
}

func viewAdminDevelopersCreateOrUpdate(ctx *fiber.Ctx) error {
	return nil
}

func viewAdminDevelopersDelete(ctx *fiber.Ctx) error {
	return nil
}

func viewAdminGamesList(ctx *fiber.Ctx) error {
	return ctx.Render("admin/games", fiber.Map{
		"Title":    "Games",
		"NavPages": getAdminPaths(GamesAdminPathGroup),
	}, "layouts/admin")
}

func viewAdminGamesRead(ctx *fiber.Ctx) error {
	return nil
}

func viewAdminGamesCreateOrUpdate(ctx *fiber.Ctx) error {
	return nil
}

func viewAdminGamesDelete(ctx *fiber.Ctx) error {
	return nil
}

func viewAdminGameAchievementsList(ctx *fiber.Ctx) error {
	return ctx.Render("admin/achievements", fiber.Map{
		"Title":    "Achievements",
		"NavPages": getAdminPaths(GamesAdminPathGroup),
	}, "layouts/admin")
}

func viewAdminGameAchievementsRead(ctx *fiber.Ctx) error {
	return nil
}

func viewAdminGameAchievementsCreateOrUpdate(ctx *fiber.Ctx) error {
	return nil
}

func viewAdminGameAchievementsDelete(ctx *fiber.Ctx) error {
	return nil
}

func SetupAdminViews(router fiber.Router) error {
	if !authSetupComplete {
		return errors.New("call SetupAuth() before calling SetupAdminViews()")
	}

	adminGroup := router.Group("/admin")
	adminGroup.Use(AuthHandler)
	adminGroup.Use(RequireAdminAuthHandler)

	adminGroup.Get("/", viewAdminHomeGet)

	adminGroup.Get("/users", viewAdminUsersList)
	adminGroup.Get("/users/:slug", viewAdminUsersRead)
	adminGroup.Put("/users/:slug", viewAdminUsersCreateOrUpdate)
	adminGroup.Delete("/users/:slug", viewAdminUsersDelete)

	adminGroup.Get("/developers", viewAdminDevelopersList)
	adminGroup.Get("/developers/:devSlug", viewAdminDevelopersRead)
	adminGroup.Put("/developers/:devSlug", viewAdminDevelopersCreateOrUpdate)
	adminGroup.Delete("/developers/:devSlug", viewAdminDevelopersDelete)

	adminGroup.Get("/developers/:devSlug/games", viewAdminGamesList)
	adminGroup.Get("/developers/:devSlug/games/:gameSlug", viewAdminGamesRead)
	adminGroup.Put("/developers/:devSlug/games/:gameSlug", viewAdminGamesCreateOrUpdate)
	adminGroup.Delete("/developers/:devSlug/games/:gameSlug", viewAdminGamesDelete)

	adminGroup.Get("/developers/:devSlug/games/:gameSlug/achievements", viewAdminGameAchievementsList)
	adminGroup.Get("/developers/:devSlug/games/:gameSlug/achievements/:achievementSlug", viewAdminGameAchievementsRead)
	adminGroup.Put("/developers/:devSlug/games/:gameSlug/achievements/:achievementSlug", viewAdminGameAchievementsCreateOrUpdate)
	adminGroup.Delete("/developers/:devSlug/games/:gameSlug/achievements/:achievementSlug", viewAdminGameAchievementsDelete)

	return nil
}

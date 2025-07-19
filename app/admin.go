package main

import (
	"context"
	"database/sql"
	"errors"
	"github.com/dresswithpockets/openstats/app/queries"
	"github.com/dresswithpockets/openstats/app/query"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

const (
	RootUserDisplayName = "Admin"
	RootUserEmail       = ""
	RootUserSlug        = "openstats"
	RootUserPass        = "openstatsadmin"
)

func AddRootAdminUser(ctx context.Context) {
	_, newUserErr := AddNewUser(ctx, RootUserDisplayName, RootUserEmail, RootUserSlug, RootUserPass)
	// this function is expected to be idempotent - if called multiple times, it shouldn't fail even if the admin
	// already exists
	if newUserErr != nil && !errors.Is(newUserErr, queries.ErrSlugAlreadyInUse) {
		log.Fatal(newUserErr)
	}
}

func IsAdmin(user *query.User) bool {
	// TODO: add distinction between Admin and Root - Root should be able to add non-root Admin users, which can do
	//       everything except add other Admins
	return user != nil && user.Slug == RootUserSlug
}

func IsRoot(user *query.User) bool {
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
	result, queryErr := Queries.AllUsersWithDisplayNames(ctx.Context())
	if queryErr != nil && !errors.Is(queryErr, sql.ErrNoRows) {
		// TODO: handle error
	}

	return ctx.Render("admin/users", fiber.Map{
		"Title":    "Users",
		"NavPages": getAdminPaths(UsersAdminPathGroup),
		"Users":    result,
	}, "layouts/admin")
}

func viewAdminUsersRead(ctx *fiber.Ctx) error {
	slug := ctx.Params("slug")
	if slug == "" || slug == "@" {
		return ctx.Redirect("/admin/users")
	}

	user, userErr := Queries.FindUserBySlug(ctx.Context(), slug)

	foundUser := true
	if errors.Is(userErr, sql.ErrNoRows) {
		ctx.Status(fiber.StatusNotFound)
		foundUser = false
	} else if userErr != nil {
		// TODO: handle error
	}

	displayNames, displayNamesErr := Queries.GetUserDisplayNames(ctx.Context(), user.ID)
	if displayNamesErr != nil && !errors.Is(displayNamesErr, sql.ErrNoRows) {
		// TODO: handle error
	}

	emails, emailsErr := Queries.GetUserEmails(ctx.Context(), user.ID)
	if emailsErr != nil && !errors.Is(emailsErr, sql.ErrNoRows) {
		// TODO: handle error
	}

	developers, developersErr := Queries.GetUserDevelopers(ctx.Context(), user.ID)
	if developersErr != nil && !errors.Is(developersErr, sql.ErrNoRows) {
		// TODO: handle error
	}

	type Model struct {
		User         query.User
		DisplayNames []query.UserDisplayName
		Emails       []query.UserEmail
		Developers   []query.GetUserDevelopersRow
	}

	return ctx.Render("admin/user", fiber.Map{
		"Title":    "User",
		"NavPages": getAdminPaths(UsersAdminPathGroup),
		"Found":    foundUser,
		"PathSlug": slug,
		"Model": &Model{
			User:         user,
			DisplayNames: displayNames,
			Emails:       emails,
			Developers:   developers,
		},
	}, "layouts/admin")
}

func viewAdminUsersCreateOrUpdate(ctx *fiber.Ctx) error {
	return nil
}

func viewAdminUsersDelete(ctx *fiber.Ctx) error {
	return nil
}

func viewAdminDevelopersList(ctx *fiber.Ctx) error {
	result, err := Queries.AllDevelopers(ctx.Context())
	if err != nil {
		// TODO: handle error
	}

	return ctx.Render("admin/developers", fiber.Map{
		"Title":      "Developers",
		"NavPages":   getAdminPaths(DevelopersAdminPathGroup),
		"Developers": result,
	}, "layouts/admin")
}

func viewAdminDevelopersRead(ctx *fiber.Ctx) error {
	slug := ctx.Params("devSlug")
	if slug == "" || slug == "@" {
		return ctx.Redirect("/admin/developers")
	}

	developer, developerErr := Queries.FindDeveloperBySlug(ctx.Context(), slug)

	foundDeveloper := true
	if errors.Is(developerErr, sql.ErrNoRows) {
		ctx.Status(fiber.StatusNotFound)
		foundDeveloper = false
	} else if developerErr != nil {
		// TODO: handle error
	}

	members, membersErr := Queries.GetDeveloperMembers(ctx.Context(), developer.ID)
	if membersErr != nil && !errors.Is(membersErr, sql.ErrNoRows) {
		// TODO: handle error
	}

	games, gamesErr := Queries.GetDeveloperGames(ctx.Context(), developer.ID)
	if gamesErr != nil && !errors.Is(gamesErr, sql.ErrNoRows) {
		// TODO: handle error
	}

	type Model struct {
		Developer query.Developer
		Members   []string
		Games     []string
	}

	return ctx.Render("admin/developer", fiber.Map{
		"Title":    "Developer",
		"NavPages": getAdminPaths(DevelopersAdminPathGroup),
		"Found":    foundDeveloper,
		"PathSlug": slug,
		"Model": &Model{
			Developer: developer,
			Members:   members,
			Games:     games,
		},
	}, "layouts/admin")
}

func viewAdminDevelopersCreateOrUpdate(ctx *fiber.Ctx) error {
	return nil
}

func viewAdminDevelopersDelete(ctx *fiber.Ctx) error {
	return nil
}

func viewAdminGamesList(ctx *fiber.Ctx) error {
	games, err := Queries.AllGames(ctx.Context())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		// TODO: handle error
	}

	return ctx.Render("admin/games", fiber.Map{
		"Title":    "Games",
		"NavPages": getAdminPaths(GamesAdminPathGroup),
		"Games":    games,
	}, "layouts/admin")
}

func viewAdminGamesRead(ctx *fiber.Ctx) error {
	devSlug := ctx.Params("devSlug")
	gameSlug := ctx.Params("gameSlug")
	if devSlug == "" || devSlug == "@" || gameSlug == "" || gameSlug == "@" {
		return ctx.Redirect("/admin/developers/@/games")
	}

	game, gameErr := Queries.FindGameBySlug(ctx.Context(), query.FindGameBySlugParams{
		DevSlug: devSlug,
		Slug:    gameSlug,
	})

	foundGame := true
	if errors.Is(gameErr, sql.ErrNoRows) {
		ctx.Status(fiber.StatusNotFound)
		foundGame = false
	} else if gameErr != nil {
		// TODO: handle error
	}

	achievements, achievementsErr := Queries.GetGameAchievements(ctx.Context(), game.ID)
	if achievementsErr != nil && !errors.Is(achievementsErr, sql.ErrNoRows) {
		// TODO: handle error
	}

	type Model struct {
		Game          query.Game
		DeveloperSlug string
		Achievements  []query.Achievement
	}

	return ctx.Render("admin/game", fiber.Map{
		"Title":    "Game",
		"NavPages": getAdminPaths(GamesAdminPathGroup),
		"Found":    foundGame,
		"Path":     ctx.Path(),
		"Model": &Model{
			Game:          game,
			DeveloperSlug: devSlug,
			Achievements:  achievements,
		},
	}, "layouts/admin")
}

func viewAdminGamesCreateOrUpdate(ctx *fiber.Ctx) error {
	return nil
}

func viewAdminGamesDelete(ctx *fiber.Ctx) error {
	return nil
}

func viewAdminGameAchievementsRead(ctx *fiber.Ctx) error {
	devSlug := ctx.Params("devSlug")
	gameSlug := ctx.Params("gameSlug")
	achievementSlug := ctx.Params("achievementSlug")
	if devSlug == "" || devSlug == "@" || gameSlug == "" || gameSlug == "@" || achievementSlug == "" || achievementSlug == "@" {
		return ctx.Redirect("/admin/developers/@/games")
	}

	achievement, achievementErr := Queries.FindAchievementBySlug(ctx.Context(), query.FindAchievementBySlugParams{
		DevSlug:  devSlug,
		GameSlug: gameSlug,
		Slug:     achievementSlug,
	})

	foundAchievement := true
	if errors.Is(achievementErr, sql.ErrNoRows) {
		ctx.Status(fiber.StatusNotFound)
		foundAchievement = false
	} else if achievementErr != nil {
		// TODO: handle error
	}

	developerPath, devRouteErr := ctx.GetRouteURL("readDeveloper", fiber.Map{"devSlug": devSlug})
	if devRouteErr != nil {
		// TODO: handle error
	}

	gamePath, gameRouteErr := ctx.GetRouteURL("readGame", fiber.Map{"devSlug": devSlug, "gameSlug": gameSlug})
	if gameRouteErr != nil {
		// TODO: handle error
	}

	type Model struct {
		Achievement   query.Achievement
		DeveloperPath string
		DeveloperSlug string
		GamePath      string
		GameSlug      string
	}

	return ctx.Render("admin/achievement", fiber.Map{
		"Title":    "Achievement",
		"NavPages": getAdminPaths(GamesAdminPathGroup),
		"Found":    foundAchievement,
		"Path":     ctx.Path(),
		"Model": &Model{
			Achievement:   achievement,
			DeveloperPath: developerPath,
			DeveloperSlug: devSlug,
			GamePath:      gamePath,
			GameSlug:      gameSlug,
		},
	}, "layouts/admin")
}

func viewAdminGameAchievementsCreateOrUpdate(ctx *fiber.Ctx) error {
	devSlug := ctx.Params("devSlug")
	gameSlug := ctx.Params("gameSlug")
	achievementSlug := ctx.Params("achievementSlug")
	if devSlug == "" || devSlug == "@" || gameSlug == "" || gameSlug == "@" || achievementSlug == "" || achievementSlug == "@" {
		return ctx.Redirect("/admin/developers/@/games")
	}

	var request struct {
		Name                string `json:"name"`
		Description         string `json:"description"`
		ProgressRequirement int64  `json:"progressRequirement"`
	}

	if bodyErr := ctx.BodyParser(&request); bodyErr != nil {
		return ctx.SendStatus(fiber.StatusBadRequest)
	}

	if len(request.Name) == 0 {
		return ctx.SendStatus(fiber.StatusBadRequest)
	}
	if request.ProgressRequirement < 0 {
		return ctx.SendStatus(fiber.StatusBadRequest)
	}

	game, gameErr := Queries.FindGameBySlug(ctx.Context(), query.FindGameBySlugParams{
		DevSlug: devSlug,
		Slug:    gameSlug,
	})

	if errors.Is(gameErr, sql.ErrNoRows) {
		return ctx.SendStatus(fiber.StatusNotFound)
	}

	if gameErr != nil {
		log.Error(gameErr)
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	isNew, createErr := Queries.UpsertAchievement(ctx.Context(), query.UpsertAchievementParams{
		GameID:              game.ID,
		Slug:                achievementSlug,
		Name:                request.Name,
		Description:         request.Description,
		ProgressRequirement: request.ProgressRequirement,
	})

	if createErr != nil {
		log.Error(createErr)
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	if isNew == 1 {
		newLocation, routeErr := ctx.GetRouteURL("readAchievement", fiber.Map{"devSlug": devSlug, "gameSlug": gameSlug, "achievementSlug": achievementSlug})
		if routeErr == nil {
			ctx.Location(newLocation)
		}
		return ctx.SendStatus(fiber.StatusCreated)
	}

	return ctx.SendStatus(fiber.StatusOK)
}

func viewAdminGameAchievementsDelete(ctx *fiber.Ctx) error {
	return nil
}

func SetupAdminViews(router fiber.Router) error {
	adminGroup := router.Group("/admin")
	adminGroup.Use(AuthHandler)
	adminGroup.Use(RequireAdminAuthHandler)

	adminGroup.Get("/", viewAdminHomeGet)

	adminGroup.Get("/users", viewAdminUsersList)
	adminGroup.Get("/users/:slug", viewAdminUsersRead)
	adminGroup.Put("/users/:slug", viewAdminUsersCreateOrUpdate)
	adminGroup.Delete("/users/:slug", viewAdminUsersDelete)

	adminGroup.Get("/developers", viewAdminDevelopersList)
	adminGroup.Get("/developers/:devSlug", viewAdminDevelopersRead).Name("readDeveloper")
	adminGroup.Put("/developers/:devSlug", viewAdminDevelopersCreateOrUpdate)
	adminGroup.Delete("/developers/:devSlug", viewAdminDevelopersDelete)

	adminGroup.Get("/developers/@/games", viewAdminGamesList)
	adminGroup.Get("/developers/:devSlug/games/:gameSlug", viewAdminGamesRead).Name("readGame")
	adminGroup.Put("/developers/:devSlug/games/:gameSlug", viewAdminGamesCreateOrUpdate)
	adminGroup.Delete("/developers/:devSlug/games/:gameSlug", viewAdminGamesDelete)

	adminGroup.Get("/developers/:devSlug/games/:gameSlug/achievements/:achievementSlug", viewAdminGameAchievementsRead)
	adminGroup.Put("/developers/:devSlug/games/:gameSlug/achievements/:achievementSlug", viewAdminGameAchievementsCreateOrUpdate)
	adminGroup.Delete("/developers/:devSlug/games/:gameSlug/achievements/:achievementSlug", viewAdminGameAchievementsDelete)

	return nil
}

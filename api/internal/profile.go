package internal

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/dresswithpockets/openstats/app/auth"
	"github.com/dresswithpockets/openstats/app/db"
	"github.com/dresswithpockets/openstats/app/db/query"
	"github.com/dresswithpockets/openstats/app/media"
	"github.com/dresswithpockets/openstats/app/rid"
	"github.com/dresswithpockets/openstats/app/users"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
)

type ProfileAchievement struct {
	Slug        string `json:"slug" readOnly:"true"`
	Name        string `json:"name" readOnly:"true"`
	Description string `json:"description" readOnly:"true"`
	AvatarUrl   string `json:"avatarUrl" readOnly:"true"`
}

type ProfileGame struct {
	RID       rid.RID `json:"rid" readOnly:"true"`
	Name      string  `json:"name" readOnly:"true"`
	AvatarUrl string  `json:"avatarUrl" readOnly:"true"`
}
type ProfileRareAchievements struct{}

type ProfileUnlockedAchievement struct {
	Game        ProfileGame `json:"game" readOnly:"true"`
	Slug        string      `json:"slug" readOnly:"true" doc:"The slug of the achievement that was unlocked"`
	Name        string      `json:"name" readOnly:"true" doc:"The name of the achievement that was unlocked"`
	Description string      `json:"description" readOnly:"true" doc:"The description of the achievement that was unlocked"`
}

type ProfileRareAchievement struct {
	Game        ProfileGame `json:"game" readOnly:"true"`
	Slug        string      `json:"slug" readOnly:"true"`
	Name        string      `json:"name" readOnly:"true"`
	Description string      `json:"description" readOnly:"true"`
	Rarity      float64     `json:"rarity" readOnly:"true" doc:"Of players who have ever played this achievement's game, the fraction who have completed this achievement"`
}

type ProfileCompletedGame struct {
	Game             ProfileGame `json:"game" readOnly:"true"`
	AchievementCount int64       `json:"achievementCount" readOnly:"true" doc:"The number of achievements necessary to complete this game"`
}

type ProfileOtherUserUnlockedAchievement struct {
	ProfileUnlockedAchievement
	User InternalUser `json:"user" readOnly:"true"`
}

type UserProfile struct {
	User InternalUser `json:"user"`

	UnlockedAchievements []ProfileUnlockedAchievement `json:"unlockedAchievements,omitempty" doc:"Most recent achievements unlocked by this user" readOnly:"true"`
	RarestAchievements   []ProfileRareAchievement     `json:"rarestAchievements,omitempty" doc:"The rarest achievements unlocked by this user" readOnly:"true"`
	CompletedGames       []ProfileCompletedGame       `json:"completedGames,omitempty" doc:"The games this user has 100% completion in" readOnly:"true"`

	// TODO: OtherUserAchievements can probably be cached with a short TTL since it'll be the same across all user profiles.
	OtherUserAchievements []ProfileOtherUserUnlockedAchievement `json:"otherUserAchievements,omitempty" doc:"Most recent achievements unlocked by other users" readOnly:"true"`
}

func GetUserProfile(ctx context.Context, userUuid uuid.UUID) (UserProfile, error) {
	sessionProfile, err := db.Queries.GetUserSessionProfile(ctx, userUuid)
	if err != nil {
		// this shouldn't ever error if the user request has as principal
		return UserProfile{}, err
	}

	recentUserAchievements, err := db.Queries.GetUserRecentAchievements(ctx, query.GetUserRecentAchievementsParams{
		UserUuid: userUuid,
		Limit:    20,
	})

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Println(err)
		return UserProfile{}, eris.Wrap(err, "couldn't get user's recent achievements")
	}

	recentOtherUserAchievements, err := db.Queries.GetOtherUserRecentAchievements(ctx, query.GetOtherUserRecentAchievementsParams{
		ExcludedUserUuid: userUuid,
		Limit:            20,
	})

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Println(err)
		return UserProfile{}, eris.Wrap(err, "couldn't get other users' recent achievements")
	}

	rarestAchievements, err := db.Queries.GetUsersRarestAchievements(ctx, query.GetUsersRarestAchievementsParams{
		UserUuid:             userUuid,
		MaxCompletionPercent: 0.1,
		Limit:                10,
	})

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Println(err)
		return UserProfile{}, eris.Wrap(err, "couldn't get user's rarest achievements")
	}

	completedGames, err := db.Queries.GetUsersCompletedGames(ctx, query.GetUsersCompletedGamesParams{
		UserUuid: userUuid,
		Limit:    5,
	})

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Println(err)
		return UserProfile{}, eris.Wrap(err, "couldn't get user's completed games")
	}

	unlocks := make([]ProfileUnlockedAchievement, len(recentUserAchievements))
	for idx, achievement := range recentUserAchievements {
		unlocks[idx] = ProfileUnlockedAchievement{
			Game: ProfileGame{
				RID:       rid.From(GameRidPrefix, achievement.GameUuid),
				Name:      achievement.GameName,
				AvatarUrl: "",
			},
			Slug:        achievement.Slug,
			Name:        achievement.Name,
			Description: achievement.Description,
		}
	}

	rarest := make([]ProfileRareAchievement, len(rarestAchievements))
	for idx, achievement := range rarestAchievements {
		rarest[idx] = ProfileRareAchievement{
			Game: ProfileGame{
				RID:       rid.From(GameRidPrefix, achievement.GameUuid),
				Name:      "", // TODO: game display names
				AvatarUrl: "", // TODO: game display avatars
			},
			Slug:        achievement.Slug,
			Name:        achievement.Name,
			Description: achievement.Description,
			Rarity:      achievement.Rarity,
		}
	}

	completed := make([]ProfileCompletedGame, len(completedGames))
	for idx, game := range completedGames {
		completed[idx] = ProfileCompletedGame{
			Game: ProfileGame{
				RID:       rid.From(GameRidPrefix, game.GameUuid),
				Name:      "", // TODO: game display names
				AvatarUrl: "", // TODO: game display avatars
			},
			AchievementCount: game.AchievementCount,
		}
	}

	otherUserUnlocks := make([]ProfileOtherUserUnlockedAchievement, len(recentUserAchievements))
	for idx, achievement := range recentOtherUserAchievements {
		otherUserUnlocks[idx] = ProfileOtherUserUnlockedAchievement{
			ProfileUnlockedAchievement: ProfileUnlockedAchievement{
				Game: ProfileGame{
					RID:       rid.From(GameRidPrefix, achievement.GameUuid),
					Name:      "", // TODO: game display names
					AvatarUrl: "", // TODO: game display avatars
				},
				Slug:        achievement.Slug,
				Name:        achievement.Name,
				Description: achievement.Description,
			},
			User: InternalUser{
				RID:         rid.From(auth.UserRidPrefix, achievement.UserUuid),
				CreatedAt:   time.Time{},
				Slug:        &achievement.UserSlug,
				DisplayName: achievement.UserDisplayName,
				BioText:     nil,
				Avatar:      nil,
			},
		}
	}

	var avatar *users.Avatar
	if sessionProfile.AvatarBlurhash != nil && sessionProfile.AvatarUuid.Valid {
		avatar = &users.Avatar{
			Url:      media.GetAvatarUrl("users", sessionProfile.AvatarUuid.UUID),
			Blurhash: *sessionProfile.AvatarBlurhash,
		}
	}

	return UserProfile{
		User: InternalUser{
			RID: rid.RID{
				Prefix: auth.UserRidPrefix,
				ID:     sessionProfile.Uuid,
			},
			CreatedAt:   sessionProfile.CreatedAt,
			Slug:        &sessionProfile.Slug,
			DisplayName: &sessionProfile.DisplayName,
			Avatar:      avatar,
		},
		UnlockedAchievements:  unlocks,
		RarestAchievements:    rarest,
		CompletedGames:        completed,
		OtherUserAchievements: otherUserUnlocks,
	}, nil
}

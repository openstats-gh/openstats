package internal

import (
	"context"
	"database/sql"
	"errors"
	"github.com/dresswithpockets/openstats/app/auth"
	"github.com/dresswithpockets/openstats/app/db"
	"github.com/dresswithpockets/openstats/app/db/query"
	"github.com/dresswithpockets/openstats/app/rid"
	"github.com/google/uuid"
	"time"
)

type GameProfileAchievement struct {
	Slug        string  `json:"slug"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Rarity      float64 `json:"rarity"`
}

type GameProfileRecentAchievements struct {
	Slug        string       `json:"slug" readOnly:"true"`
	Name        string       `json:"name" readOnly:"true"`
	Description string       `json:"description" readOnly:"true"`
	Rarity      float64      `json:"rarity" readOnly:"true"`
	User        InternalUser `json:"user" readOnly:"true"`
}

type GameProfileRecentCompletionists struct {
	User       InternalUser `json:"user" readOnly:"true"`
	UnlockedAt time.Time    `json:"unlockedAt" readOnly:"true"`
}

type GameProfile struct {
	Game InternalGame `json:"game"`

	Achievements         []GameProfileAchievement          `json:"achievements,omitempty"`
	RecentAchievements   []GameProfileRecentAchievements   `json:"recentAchievements,omitempty"`
	RecentCompletionists []GameProfileRecentCompletionists `json:"recentCompletionists,omitempty"`
}

// GetGameProfile returns all information necessary to render a specific game's profile
func GetGameProfile(ctx context.Context, gameUuid uuid.UUID) (*GameProfile, error) {
	gameProfile, err := db.Queries.GetGameProfile(ctx, gameUuid)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	gameAchievements, err := db.Queries.GetGameAchievementsWithRarity(ctx, gameUuid)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	recentGameAchievements, err := db.Queries.GetRecentGameAchievements(ctx, query.GetRecentGameAchievementsParams{
		GameUuid: gameUuid,
		Limit:    20,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	recentGameCompletions, err := db.Queries.GetRecentGameCompletions(ctx, query.GetRecentGameCompletionsParams{
		GameUuid: gameUuid,
		Limit:    20,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	achievements := make([]GameProfileAchievement, len(gameAchievements))
	for idx, achievement := range gameAchievements {
		achievements[idx] = GameProfileAchievement{
			Slug:        achievement.Slug,
			Name:        achievement.Name,
			Description: achievement.Description,
			Rarity:      achievement.Rarity,
		}
	}

	recentAchievements := make([]GameProfileRecentAchievements, len(recentGameAchievements))
	for idx, recentAchievement := range recentGameAchievements {
		recentAchievements[idx] = GameProfileRecentAchievements{
			Slug:        recentAchievement.Slug,
			Name:        recentAchievement.Name,
			Description: recentAchievement.Description,
			Rarity:      recentAchievement.Rarity,
			User: InternalUser{
				RID:  rid.From(auth.UserRidPrefix, recentAchievement.UserUuid),
				Slug: &recentAchievement.UserSlug,
			},
		}
	}

	recentCompletionists := make([]GameProfileRecentCompletionists, len(recentGameCompletions))
	for idx, recentCompletion := range recentGameCompletions {
		recentCompletionists[idx] = GameProfileRecentCompletionists{
			User: InternalUser{
				RID:  rid.From(auth.UserRidPrefix, recentCompletion.UserUuid),
				Slug: &recentCompletion.UserSlug,
			},
			UnlockedAt: recentCompletion.UnlockedAt,
		}
	}

	return &GameProfile{
		Game: InternalGame{
			RID: rid.From(GameRidPrefix, gameUuid),
			Developer: Developer{
				FriendlyName: gameProfile.DeveloperSlug,
			},
			FriendlyName: gameProfile.Slug,
		},
		Achievements:         achievements,
		RecentAchievements:   recentAchievements,
		RecentCompletionists: recentCompletionists,
	}, nil
}

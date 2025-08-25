package internal

import (
	"context"
	"database/sql"
	"errors"
	"github.com/dresswithpockets/openstats/app/auth"
	"github.com/dresswithpockets/openstats/app/db"
	"github.com/dresswithpockets/openstats/app/db/query"
	"github.com/dresswithpockets/openstats/app/media"
	"github.com/dresswithpockets/openstats/app/rid"
	"github.com/dresswithpockets/openstats/app/users"
	"github.com/google/uuid"
	"log"
)

type UserProfile struct {
	User                 InternalUser                  `json:"user"`
	UnlockedAchievements []ProfileUnlockedAchievements `json:"unlockedAchievements,omitempty" doc:"Most recent achievements unlocked by this user" readOnly:"true"`
	// TODO: OtherUserAchievements can probably be cached with a short TTL since it'll be the same across all user profiles.
	OtherUserAchievements []ProfileOtherUserUnlockedAchievements `json:"otherUserAchievements,omitempty" doc:"Most recent achievements unlocked by other users" readOnly:"true"`
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
		return UserProfile{}, err
	}

	recentOtherUserAchievements, err := db.Queries.GetOtherUserRecentAchievements(ctx, query.GetOtherUserRecentAchievementsParams{
		ExcludedUserUuid: userUuid,
		Limit:            20,
	})

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Println(err)
		return UserProfile{}, err
	}

	unlocks := make([]ProfileUnlockedAchievements, len(recentUserAchievements))
	for unlockIdx, _ := range unlocks {
		unlocks[unlockIdx].MapFromRow(recentUserAchievements[unlockIdx])
	}

	otherUserUnlocks := make([]ProfileOtherUserUnlockedAchievements, len(recentUserAchievements))
	for unlockIdx, _ := range otherUserUnlocks {
		otherUserUnlocks[unlockIdx].MapFromRow(recentOtherUserAchievements[unlockIdx])
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
		OtherUserAchievements: otherUserUnlocks,
	}, nil
}

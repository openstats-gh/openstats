package main

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/dresswithpockets/openstats/app/db/query"
)

type UnlockedAchievementInfo struct {
	DeveloperSlug string
	GameSlug      string
	GameName      string
	Slug          string
	Name          string
	Description   string
}

type OtherUserUnlockedAchievementInfo struct {
	DeveloperSlug   string
	GameSlug        string
	GameName        string
	Slug            string
	Name            string
	Description     string
	UserSlug        string
	UserDisplayName string
}

type UserBriefBody struct {
	Unlocks          []UnlockedAchievementInfo
	OtherUserUnlocks []OtherUserUnlockedAchievementInfo
}

type UserBriefResponse struct {
	Body UserBriefBody
}

type UserBriefRequest struct {
	Slug string `path:"slug" required:"true" pattern:"[a-z0-9-]+" patternDescription:"lowercase-alphanum with dashes" minLength:"2" maxLength:"64"`
}

func HandleGetUsersBrief(ctx context.Context, input *UserBriefRequest) (*UserBriefResponse, error) {
	recentUserAchievements, err := Queries.GetUserRecentAchievements(ctx, query.GetUserRecentAchievementsParams{
		UserSlug: input.Slug,
		Limit:    20,
	})

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Println(err)
		return nil, err
	}

	recentOtherUserAchievements, err := Queries.GetOtherUserRecentAchievements(ctx, query.GetOtherUserRecentAchievementsParams{
		ExcludedUserSlug: input.Slug,
		Limit:            20,
	})

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Println(err)
		return nil, err
	}

	var unlocks []UnlockedAchievementInfo
	for _, row := range recentUserAchievements {
		unlocks = append(unlocks, UnlockedAchievementInfo{
			DeveloperSlug: row.DeveloperSlug,
			GameSlug:      row.GameSlug,
			GameName:      row.GameName,
			Slug:          row.Slug,
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
			Slug:            row.Slug,
			Name:            row.Name,
			Description:     row.Description,
			UserSlug:        row.UserSlug,
			UserDisplayName: userDisplayName,
		})
	}

	return &UserBriefResponse{
		Body: UserBriefBody{
			Unlocks:          unlocks,
			OtherUserUnlocks: otherUserUnlocks,
		},
	}, nil
}

package users

import (
	"context"
	"database/sql"
	"errors"
	"github.com/danielgtaylor/huma/v2"
	"github.com/dresswithpockets/openstats/app/auth"
	"github.com/dresswithpockets/openstats/app/db"
	"github.com/dresswithpockets/openstats/app/db/query"
	"github.com/dresswithpockets/openstats/app/validation"
	"github.com/rotisserie/eris"
	"log"
	"net/http"
	"time"
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
	recentUserAchievements, err := db.Queries.GetUserRecentAchievements(ctx, query.GetUserRecentAchievementsParams{
		UserSlug: input.Slug,
		Limit:    20,
	})

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Println(err)
		return nil, err
	}

	recentOtherUserAchievements, err := db.Queries.GetOtherUserRecentAchievements(ctx, query.GetOtherUserRecentAchievementsParams{
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

type ListUsersRequest struct {
	Slug  validation.Optional[string]              `query:"slug"`
	Limit validation.Optional[int]                 `query:"limit" minimum:"10" maximum:"50" doc:"default = 10"`
	After validation.Optional[validation.LookupID] `query:"after" format:"uuid"`
}

type ListUsersItem struct {
	LookupID    validation.LookupID `json:"lookupId"`
	Slug        string              `json:"slug"`
	DisplayName string              `json:"displayName"`
	CreatedAt   time.Time           `json:"createdAt"`
}

type ListUsersBody struct {
	Items []ListUsersItem `json:"items"`
}

type ListUsersResponse struct {
	Body ListUsersBody
}

func HandleListUsers(ctx context.Context, input *ListUsersRequest) (*ListUsersResponse, error) {
	builder := db.DB.Builder().
		Select("u.lookup_id", "u.slug", "uldn.display_name", "u.created_at").
		From("users u").
		JoinClause("left outer join user_latest_display_name uldn on u.id = uldn.user_id").
		OrderBy("u.lookup_id desc")

	if input.Slug.HasValue {
		builder = builder.Where("u.slug like ?", "%"+input.Slug.Value+"%")
	}

	if input.After.HasValue {
		builder = builder.Where("u.lookup_id > ?", input.After.Value)
	}

	limit := input.Limit.ValueOr(10)
	builder = builder.Limit(uint64(limit))

	rows, queryErr := db.DB.Query(ctx, builder)
	if queryErr != nil {
		return nil, eris.Wrap(queryErr, "")
	}

	defer rows.Close()

	var items []ListUsersItem
	for rows.Next() {
		var item ListUsersItem
		if scanErr := rows.Scan(&item.LookupID, &item.Slug, &item.DisplayName, &item.CreatedAt); scanErr != nil {
			return nil, eris.Wrap(scanErr, "")
		}

		items = append(items, item)
	}

	return &ListUsersResponse{
		Body: ListUsersBody{
			Items: items,
		},
	}, nil
}

func HandlerTODO(ctx context.Context, input *struct{}) (*struct{}, error) {
	panic("HandlerTODO")
}

func RegisterRoutes(api huma.API) {
	userApi := huma.NewGroup(api, "/users/v1")
	userApi.UseSimpleModifier(func(op *huma.Operation) {
		op.Tags = append(op.Tags, "Users")
	})
	userApi.UseMiddleware(auth.UserAuthHandler)

	requireUserHandler := auth.CreateRequireUserAuthHandler(userApi)

	huma.Register(userApi, huma.Operation{
		OperationID: "get-users-brief",
		Method:      http.MethodGet,
		Path:        "/{slug}/brief",
		Summary:     "Get user brief",
		Description: "Get a detail summary containing the user's recent achievements, for display",
	}, HandleGetUsersBrief)

	huma.Register(userApi, huma.Operation{
		OperationID: "list-users",
		Method:      http.MethodGet,
		Path:        "/",
		Summary:     "List users",
		Description: "Query & filter all users",
	}, HandleListUsers)

	huma.Register(userApi, huma.Operation{
		OperationID: "create-users",
		Method:      http.MethodPost,
		Path:        "/",
		Summary:     "Create new users",
		Description: "Create 1 or more users. Requires an admin session.",
		Security:    []map[string][]string{{"Cookie": {}}},
		Middlewares: huma.Middlewares{requireUserHandler},
		Errors: []int{
			http.StatusUnauthorized,
		},
	}, HandlerTODO)

	huma.Register(userApi, huma.Operation{
		OperationID: "read-user",
		Method:      http.MethodGet,
		Path:        "/{slug}",
		Summary:     "Read user",
		Description: "Get some details for a particular user",
	}, HandlerTODO)

	huma.Register(userApi, huma.Operation{
		OperationID: "upsert-user",
		Method:      http.MethodPut,
		Path:        "/{slug}",
		Summary:     "Create or update user",
		Description: "Create or update a user at the slug specified. This is an upsert operation - it will try to create the user if it doesn't already exist, and will otherwise update an existing user.",
		Security:    []map[string][]string{{"Cookie": {}}},
		Middlewares: huma.Middlewares{requireUserHandler},
		Errors: []int{
			http.StatusUnauthorized,
		},
	}, HandlerTODO)

	huma.Register(userApi, huma.Operation{
		OperationID: "patch-user",
		Method:      http.MethodPatch,
		Path:        "/{slug}",
		Summary:     "Update user",
		Description: "Update an existing user at the slug.",
		Security:    []map[string][]string{{"Cookie": {}}},
		Middlewares: huma.Middlewares{requireUserHandler},
		Errors: []int{
			http.StatusUnauthorized,
		},
	}, HandlerTODO)

	huma.Register(userApi, huma.Operation{
		OperationID: "delete-user",
		Method:      http.MethodDelete,
		Path:        "/{slug}",
		Summary:     "Delete a user",
		Description: "Delete an existing user at the slug. Must be an Admin.",
		Security:    []map[string][]string{{"Cookie": {}}},
		Middlewares: huma.Middlewares{requireUserHandler},
		Errors: []int{
			http.StatusUnauthorized,
		},
	}, HandlerTODO)
}

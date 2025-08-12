package users

import (
	"context"
	"database/sql"
	"errors"
	"github.com/danielgtaylor/huma/v2"
	"github.com/dresswithpockets/openstats/app/auth"
	"github.com/dresswithpockets/openstats/app/db"
	"github.com/dresswithpockets/openstats/app/db/query"
	"github.com/dresswithpockets/openstats/app/rid"
	"github.com/dresswithpockets/openstats/app/validation"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rotisserie/eris"
	"log"
	"net/http"
	"time"
)

// User resource returned by users/ endpoints
type User struct {
	RID         rid.RID   `json:"rid" readOnly:"true"`
	CreatedAt   time.Time `json:"createdAt" readOnly:"true"`
	Slug        string    `json:"slug"`
	DisplayName string    `json:"displayName,omitempty"`
	BioText     string    `json:"bioText,omitempty"`
	AvatarUrl   string    `json:"avatarUrl,omitempty" readOnly:"true"`
	Email       string    `json:"email,omitempty"`
	Password    string    `json:"password,omitempty" writeOnly:"true"`
}

type UnlockedAchievementInfo struct {
	DeveloperSlug string
	GameSlug      string
	GameName      string
	Slug          string
	Name          string
	Description   string
}

type OtherUserUnlockedAchievementInfo struct {
	DeveloperSlug    string
	GameSlug         string
	GameName         string
	Slug             string
	Name             string
	Description      string
	UserRID          rid.RID
	UserFriendlyName string
}

type UserBrief struct {
	Unlocks          []UnlockedAchievementInfo          `json:"unlocks"`
	OtherUserUnlocks []OtherUserUnlockedAchievementInfo `json:"otherUserUnlocks"`
}

type UserBriefBody struct {
	UserBrief UserBrief `json:"userBrief"`
}

type UserBriefResponse struct {
	Body UserBriefBody
}

type UserBriefRequest struct {
	UserId validation.SlugOrRID `path:"userId" required:"true"`
}

func HandleGetUsersBrief(ctx context.Context, input *UserBriefRequest) (*UserBriefResponse, error) {
	userRID, ensureErr := validation.EnsureRID(ctx, input.UserId, auth.UserRidPrefix, db.Queries.GetUserUuid)
	if ensureErr != nil {
		return nil, ensureErr
	}

	recentUserAchievements, err := db.Queries.GetUserRecentAchievements(ctx, query.GetUserRecentAchievementsParams{
		UserUuid: userRID.ID,
		Limit:    20,
	})

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Println(err)
		return nil, err
	}

	recentOtherUserAchievements, err := db.Queries.GetOtherUserRecentAchievements(ctx, query.GetOtherUserRecentAchievementsParams{
		ExcludedUserUuid: userRID.ID,
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
		otherUserUnlocks = append(otherUserUnlocks, OtherUserUnlockedAchievementInfo{
			DeveloperSlug: row.DeveloperSlug,
			GameSlug:      row.GameSlug,
			GameName:      row.GameName,
			Slug:          row.Slug,
			Name:          row.Name,
			Description:   row.Description,
			UserRID: rid.RID{
				Prefix: auth.UserRidPrefix,
				ID:     row.UserUuid,
			},
			UserFriendlyName: row.UserFriendlyName,
		})
	}

	return &UserBriefResponse{
		Body: UserBriefBody{
			UserBrief: UserBrief{
				Unlocks:          unlocks,
				OtherUserUnlocks: otherUserUnlocks,
			},
		},
	}, nil
}

type ListUsersRequest struct {
	// TODO: oneOf Slug and SlugContains
	Slug         validation.Optional[string]  `query:"slug"`
	SlugContains validation.Optional[string]  `query:"slugContains"`
	Limit        validation.Optional[int]     `query:"limit" minimum:"10" maximum:"50" doc:"default = 10"`
	After        validation.Optional[rid.RID] `query:"after"`
}

type ListUsersBody struct {
	Users []User `json:"users"`
}

type ListUsersResponse struct {
	Body ListUsersBody
}

func HandleListUsers(ctx context.Context, input *ListUsersRequest) (*ListUsersResponse, error) {
	principal, hasPrincipal := auth.GetPrincipal(ctx)
	isAdmin := hasPrincipal && auth.IsAdmin(principal.User)

	var principalUuid uuid.UUID
	if hasPrincipal {
		principalUuid = principal.User.Uuid
	}

	builder := db.DB.Builder().
		Select("u.uuid", "u.slug", "coalesce(uldn.display_name, '')", "coalesce(ule.email, '')", "u.created_at").
		From("users u").
		JoinClause("left outer join user_latest_display_name uldn on u.id = uldn.user_id").
		JoinClause("left outer join user_latest_email ule on u.id = ule.user_id and (? or (? and u.uuid = ?))", isAdmin, hasPrincipal, principalUuid).
		OrderBy("u.uuid desc")

	if input.Slug.HasValue {
		builder = builder.Where("u.slug = ?", input.Slug.Value)
	} else if input.SlugContains.HasValue {
		builder = builder.Where("u.slug like ?", "%"+input.SlugContains.Value+"%")
	}

	if input.After.HasValue {
		builder = builder.Where("u.rid > ?", input.After.Value)
	}

	limit := input.Limit.ValueOr(10)
	builder = builder.Limit(uint64(limit))

	rows, queryErr := db.DB.Query(ctx, builder)
	if queryErr != nil && !errors.Is(queryErr, sql.ErrNoRows) {
		return nil, eris.Wrap(queryErr, "")
	}

	defer rows.Close()

	var users []User
	for rows.Next() {
		var userUuid uuid.UUID
		var item User

		if scanErr := rows.Scan(
			&userUuid,
			&item.Slug,
			&item.DisplayName,
			&item.Email,
			&item.CreatedAt,
		); scanErr != nil {
			return nil, eris.Wrap(scanErr, "")
		}

		item.RID = rid.RID{
			Prefix: auth.UserRidPrefix,
			ID:     userUuid,
		}
		users = append(users, item)
	}

	return &ListUsersResponse{
		Body: ListUsersBody{
			Users: users,
		},
	}, nil
}

type ReadUserRequest struct {
	UserId validation.SlugOrRID `path:"userId" required:"true"`
}

type ReadUserResponseBody struct {
	User User `json:"user"`
}

type ReadUserResponse struct {
	Body ReadUserResponseBody
}

func HandleReadUser(ctx context.Context, input *ReadUserRequest) (*ReadUserResponse, error) {
	principal, hasPrincipal := auth.GetPrincipal(ctx)
	isAdmin := hasPrincipal && auth.IsAdmin(principal.User)

	var principalUuid uuid.UUID
	if hasPrincipal {
		principalUuid = principal.User.Uuid
	}

	builder := db.DB.Builder().
		Select("u.uuid", "u.slug", "coalesce(uldn.display_name, '')", "coalesce(ule.email, '')", "u.created_at").
		From("users u").
		JoinClause("left outer join user_latest_display_name uldn on u.id = uldn.user_id").
		JoinClause("left outer join user_latest_email ule on u.id = ule.user_id and (? or (? and u.uuid = ?))", isAdmin, hasPrincipal, principalUuid)

	if userSlug, isSlug := input.UserId.Slug(); isSlug {
		builder = builder.Where("u.slug = ?", userSlug)
	} else {
		ridValue, _ := input.UserId.RID()
		builder = builder.Where("u.uuid = ?", ridValue.ID)
	}

	result := User{
		RID: rid.RID{Prefix: auth.UserRidPrefix},
	}
	scanErr := db.DB.ScanRow(
		ctx,
		builder,
		&result.RID.ID,
		&result.Slug,
		&result.DisplayName,
		&result.Email,
		&result.CreatedAt,
	)
	if errors.Is(scanErr, pgx.ErrNoRows) {
		return nil, huma.Error404NotFound("no user with matching slug")
	}
	if scanErr != nil {
		return nil, eris.Wrap(scanErr, "")
	}

	return &ReadUserResponse{
		Body: ReadUserResponseBody{
			User: result,
		},
	}, nil
}

type PutUserRequest struct {
	User User `json:"user"`
}

type PutUserResponseBody struct {
	User User `json:"user"`
}

type PutUserResponse struct {
	Status int
	Body   PutUserResponseBody
}

func HandlePutUser(ctx context.Context, input *PutUserRequest) (*PutUserResponse, error) {
	principal, hasPrincipal := auth.GetPrincipal(ctx)
	if !hasPrincipal {
		return nil, huma.Error401Unauthorized("a session is required")
	}

	// TODO: let Root user flag users as Admin
	if !auth.IsAdmin(principal.User) {
		return nil, huma.Error401Unauthorized("you are not authorized to create this user")
	}

	newUser, err := auth.AddNewUser(ctx, input.User.DisplayName, input.User.Email, input.User.Slug, input.User.Password)
	if errors.Is(err, db.ErrSlugAlreadyInUse) {
		// TODO: better conflict mechanism (take advantage of problem details, huma.ErrorDetailer?)
		return nil, huma.Error409Conflict("slug is already in use")
	}

	if err != nil {
		return nil, err
	}

	return &PutUserResponse{
		Status: http.StatusCreated,
		Body: PutUserResponseBody{
			User: User{
				RID:         rid.RID{Prefix: auth.UserRidPrefix, ID: newUser.Uuid},
				Slug:        newUser.Slug,
				DisplayName: input.User.DisplayName,
				Email:       input.User.Email,
				CreatedAt:   newUser.CreatedAt,
			},
		},
	}, nil
}

func RegisterRoutesOld(api huma.API) {
	userApi := huma.NewGroup(api, "/users/v1")
	userApi.UseSimpleModifier(func(op *huma.Operation) {
		op.Tags = append(op.Tags, "Users")
		// Until we support API keys, this API is never publicly accessible. See UseMiddleware note below.
		op.Security = []map[string][]string{{"SessionCookie": {}}}
	})
	userApi.UseMiddleware(auth.UserAuthHandler)

	requireUserHandler := auth.CreateRequireUserAuthHandler(userApi)
	// N.B. until we support API keys, this API is never publicly accessible. See op.Security usage in simple modifier
	//      above
	userApi.UseMiddleware(requireUserHandler)

	// TODO: support prefix_UID and slug interop
	//       Resources which have a UUID can be referenced by the RID format: prefix_UUID
	//       Resources which have a slug can be referenced by their slug
	//       slugs and prefix_UUID are always distinguishable, because slugs cannot contain underscores
	//       .
	//       e.g. User can be referenced via u_20W9MCAgSo06z or my-user-slug
	//       .
	//       N.B the UUIDs are encoded as base62
	//       RIDs are non-fungible - once allocated for a Resource they will never be reused for another of the same kind of Resource, and that RID will always point to that Resource
	//       slugs are mutable - a slug may only point to a single Resource at a time, but it can be changed. Slugs may be reused by other instances of the same kind of Resource if that slug isn't currently in use.

	huma.Register(userApi, huma.Operation{
		OperationID: "get-users-brief",
		Method:      http.MethodGet,
		Path:        "/{userId}/brief",
		Summary:     "Get user brief",
		Description: "Get a detail summary containing the user's recent achievements, for display",
		Security:    []map[string][]string{{"SessionCookie": {}}},
	}, HandleGetUsersBrief)

	huma.Register(userApi, huma.Operation{
		OperationID: "list-users",
		Method:      http.MethodGet,
		Path:        "/",
		Summary:     "List users",
		Description: "Query & filter all users",
		Security:    []map[string][]string{{"SessionCookie": {}}},
	}, HandleListUsers)

	huma.Register(userApi, huma.Operation{
		OperationID: "read-user",
		Method:      http.MethodGet,
		Path:        "/{userId}",
		Summary:     "Read a user",
		Description: "Get some details for a particular user",
		Security:    []map[string][]string{{"SessionCookie": {}}},
	}, HandleReadUser)

	huma.Register(userApi, huma.Operation{
		OperationID: "put-user",
		Method:      http.MethodPut,
		Path:        "/{userId}",
		Summary:     "Create a user",
		Description: "Create a user at the slug specified.",
		Security:    []map[string][]string{{"SessionCookie": {}}},
		Errors: []int{
			http.StatusUnauthorized,
		},
	}, HandlePutUser)

	huma.Register(userApi, huma.Operation{
		OperationID: "patch-user",
		Method:      http.MethodPatch,
		Path:        "/{userId}",
		Summary:     "Update a user",
		Description: "Update an existing user at the slug.",
		Security:    []map[string][]string{{"SessionCookie": {}}},
		Errors: []int{
			http.StatusUnauthorized,
		},
	}, HandlerTODO)

	huma.Register(userApi, huma.Operation{
		OperationID: "delete-user",
		Method:      http.MethodDelete,
		Path:        "/{userId}",
		Summary:     "Delete a user",
		Description: "Delete an existing user at the slug. Must be an Admin.",
		Security:    []map[string][]string{{"SessionCookie": {}}},
		Errors: []int{
			http.StatusUnauthorized,
		},
	}, HandlerTODO)
}

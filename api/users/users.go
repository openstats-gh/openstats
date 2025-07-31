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
	"github.com/jackc/pgx/v5"
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

// User resource returned by users/ endpoints
type User struct {
	LookupID    validation.LookupID `json:"lookupId"`
	Slug        string              `json:"slug"`
	DisplayName string              `json:"displayName,omitempty"`
	Email       string              `json:"email,omitempty"`
	CreatedAt   time.Time           `json:"createdAt"`
}

type CreateUser struct {
	Slug        string `json:"slug"`
	Password    string `json:"password"`
	DisplayName string `json:"displayName,omitempty"`
	Email       string `json:"email,omitempty"`
}

type MutateUser struct {
	Slug        string `json:"slug,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	Email       string `json:"email,omitempty"`
	Password    string `json:"password,omitempty"`
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
	// TODO: oneOf Slug and SlugContains
	Slug         validation.Optional[string]              `query:"slug"`
	SlugContains validation.Optional[string]              `query:"slugContains"`
	Limit        validation.Optional[int]                 `query:"limit" minimum:"10" maximum:"50" doc:"default = 10"`
	After        validation.Optional[validation.LookupID] `query:"after" format:"uuid"`
}

type ListUsersBody struct {
	Users []User `json:"users"`
}

type ListUsersResponse struct {
	Body ListUsersBody
}

func HandleListUsers(ctx context.Context, input *ListUsersRequest) (*ListUsersResponse, error) {
	// TODO: only include displayName and email upon request
	// TODO: limit email access to Admin or authenticated user matching user in query
	builder := db.DB.Builder().
		Select("u.lookup_id", "u.slug", "coalesce(uldn.display_name, '')", "coalesce(ule.email, '')", "u.created_at").
		From("users u").
		JoinClause("left outer join user_latest_display_name uldn on u.id = uldn.user_id").
		JoinClause("left outer join user_latest_email ule on u.id = ule.user_id").
		OrderBy("u.lookup_id desc")

	if input.Slug.HasValue {
		builder = builder.Where("u.slug = ?", input.Slug.Value)
	} else if input.SlugContains.HasValue {
		builder = builder.Where("u.slug like ?", "%"+input.SlugContains.Value+"%")
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

	var users []User
	for rows.Next() {
		var item User
		if scanErr := rows.Scan(
			&item.LookupID,
			&item.Slug,
			&item.DisplayName,
			&item.Email,
			&item.CreatedAt,
		); scanErr != nil {
			return nil, eris.Wrap(scanErr, "")
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
	Slug string `path:"slug"`
}

type ReadUserResponse struct {
	Body struct {
		User User `json:"user"`
	}
}

func HandleReadUser(ctx context.Context, input *ReadUserRequest) (*ReadUserResponse, error) {
	// TODO: only include displayName and email upon request
	// TODO: limit email access to Admin or authenticated user matching user in query
	builder := db.DB.Builder().
		Select("u.lookup_id", "u.slug", "coalesce(uldn.display_name, '')", "coalesce(ule.email, '')", "u.created_at").
		From("users u").
		JoinClause("left outer join user_latest_display_name uldn on u.id = uldn.user_id").
		JoinClause("left outer join user_latest_email ule on u.id = ule.user_id").
		Where("u.slug = ?", input.Slug)

	var result ReadUserResponse
	scanErr := db.DB.ScanRow(
		ctx,
		builder,
		&result.Body.User.LookupID,
		&result.Body.User.Slug,
		&result.Body.User.DisplayName,
		&result.Body.User.Email,
		&result.Body.User.CreatedAt,
	)
	if errors.Is(scanErr, pgx.ErrNoRows) {
		return nil, huma.Error404NotFound("no user with matching slug")
	}
	if scanErr != nil {
		return nil, eris.Wrap(scanErr, "")
	}

	return &result, nil
}

type PutUserRequest struct {
	User CreateUser `json:"user"`
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
				LookupID:    validation.LookupID(newUser.LookupID),
				Slug:        newUser.Slug,
				DisplayName: input.User.DisplayName,
				Email:       input.User.Email,
				CreatedAt:   newUser.CreatedAt,
			},
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
		OperationID: "read-user",
		Method:      http.MethodGet,
		Path:        "/{slug}",
		Summary:     "Read a user",
		Description: "Get some details for a particular user",
	}, HandleReadUser)

	huma.Register(userApi, huma.Operation{
		OperationID: "put-user",
		Method:      http.MethodPut,
		Path:        "/{slug}",
		Summary:     "Create a user",
		Description: "Create a user at the slug specified.",
		Security:    []map[string][]string{{"Cookie": {}}},
		Middlewares: huma.Middlewares{requireUserHandler},
		Errors: []int{
			http.StatusUnauthorized,
		},
	}, HandlePutUser)

	huma.Register(userApi, huma.Operation{
		OperationID: "patch-user",
		Method:      http.MethodPatch,
		Path:        "/{slug}",
		Summary:     "Update a user",
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

package internal

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
	"github.com/rotisserie/eris"
	"log"
	"net/http"
	"time"
)

func RegisterRoutes(api huma.API) {
	internalApi := huma.NewGroup(api, "/internal")

	requireUserAuthHandler := auth.CreateRequireUserAuthHandler(internalApi)
	internalApi.UseSimpleModifier(func(op *huma.Operation) {
		if _, noAuth := op.Metadata["NoUserAuth"]; !noAuth {
			op.Security = append(op.Security, map[string][]string{"SessionCookie": {}})
			op.Middlewares = append(op.Middlewares, auth.UserAuthHandler, requireUserAuthHandler)
			op.Errors = append(op.Errors, http.StatusUnauthorized)
		}
	})

	sessionApi := huma.NewGroup(internalApi, "/session")
	sessionApi.UseSimpleModifier(func(op *huma.Operation) {
		op.Tags = append(op.Tags, "Internal/Session")
	})
	huma.Register(sessionApi, huma.Operation{
		Path:        "/sign-up",
		OperationID: "internal-a-sign-up",
		Method:      http.MethodPost,
		Errors: []int{
			http.StatusUnauthorized,
			http.StatusConflict,
		},
		Metadata:    map[string]any{"NoUserAuth": true},
		Summary:     "Sign up",
		Description: "Create a new user and sign into a new session as the new user",
	}, auth.HandlePostSignUp)

	huma.Register(sessionApi, huma.Operation{
		Path:        "/sign-in",
		OperationID: "internal-b-sign-in",
		Method:      http.MethodPost,
		Errors:      []int{http.StatusUnauthorized},
		Metadata:    map[string]any{"NoUserAuth": true},
		Summary:     "Sign in",
		Description: "Sign into a new session as an existing user",
	}, auth.HandlePostSignIn)

	huma.Register(sessionApi, huma.Operation{
		Path:        "/sign-out",
		OperationID: "internal-c-sign-out",
		Method:      http.MethodPost,
		Middlewares: huma.Middlewares{auth.UserAuthHandler, requireUserAuthHandler}, // TODO: https://github.com/danielgtaylor/huma/issues/804
		Summary:     "Sign out",
		Description: "Sign out of the current session, and invalidate the session token",
	}, auth.HandlePostSignOut)

	huma.Register(sessionApi, huma.Operation{
		Path:        "/",
		OperationID: "internal-d-get-session",
		Method:      http.MethodGet,
		Middlewares: huma.Middlewares{auth.UserAuthHandler, requireUserAuthHandler}, // TODO: https://github.com/danielgtaylor/huma/issues/804
		Summary:     "Get session summary",
		Description: "Get details about the current authenticated session and the associated user",
	}, auth.HandleGetSession)

	huma.Register(sessionApi, huma.Operation{
		Path:        "/profile",
		OperationID: "internal-e-get-session-profile",
		Method:      http.MethodGet,
		Middlewares: huma.Middlewares{auth.UserAuthHandler, requireUserAuthHandler}, // TODO: https://github.com/danielgtaylor/huma/issues/804
		Summary:     "Get user's profile",
		Description: "Get profile of current authenticated user",
	}, HandleGetSessionProfile)

	huma.Register(sessionApi, huma.Operation{
		Path:        "/profile",
		OperationID: "internal-f-update-session-profile",
		Method:      http.MethodPost,
		Middlewares: huma.Middlewares{auth.UserAuthHandler, requireUserAuthHandler}, // TODO: https://github.com/danielgtaylor/huma/issues/804
		Summary:     "Update user's profile",
		Description: "Update profile of current authenticated user",
	}, HandlePostSessionProfile)

	huma.Register(sessionApi, huma.Operation{
		Path:        "/tokens",
		OperationID: "internal-f-get-tokens",
		Method:      http.MethodGet,
		Middlewares: huma.Middlewares{auth.UserAuthHandler, requireUserAuthHandler}, // TODO: https://github.com/danielgtaylor/huma/issues/804
		Summary:     "Get user's tokens",
		Description: "Get all of the current user's tokens",
	}, HandleGetSessionGameTokens)

	huma.Register(sessionApi, huma.Operation{
		Path:        "/tokens",
		OperationID: "internal-g-create-token",
		Method:      http.MethodPost,
		Errors:      []int{http.StatusBadRequest},
		Middlewares: huma.Middlewares{auth.UserAuthHandler, requireUserAuthHandler}, // TODO: https://github.com/danielgtaylor/huma/issues/804
		Summary:     "Create a new token",
		Description: "Create a new token for the current user",
	}, HandlePostSessionGameToken)

	huma.Register(sessionApi, huma.Operation{
		Path:        "/tokens/{tokenRID}",
		OperationID: "internal-h-delete-token",
		Method:      http.MethodDelete,
		Errors:      []int{http.StatusBadRequest},
		Middlewares: huma.Middlewares{auth.UserAuthHandler, requireUserAuthHandler}, // TODO: https://github.com/danielgtaylor/huma/issues/804
		Summary:     "Invalidate a token",
		Description: "Invalidate one of the current user's tokens",
	}, HandleDeleteSessionGameToken)

	userApi := huma.NewGroup(internalApi, "/users")
	userApi.UseSimpleModifier(func(op *huma.Operation) {
		op.Tags = append(op.Tags, "Internal/Users")
	})
	huma.Register(userApi, huma.Operation{
		Path:        "/users",
		OperationID: "internal-i-search-users",
		Method:      http.MethodGet,
		Middlewares: huma.Middlewares{auth.UserAuthHandler, requireUserAuthHandler}, // TODO: https://github.com/danielgtaylor/huma/issues/804
		Summary:     "Search users",
		Description: "Search all users by various criteria",
	}, HandleSearchUsers)

	huma.Register(userApi, huma.Operation{
		Path:        "/users/{user}/profile",
		OperationID: "internal-h-get-user-profile",
		Method:      http.MethodGet,
		Middlewares: huma.Middlewares{auth.UserAuthHandler, requireUserAuthHandler}, // TODO: https://github.com/danielgtaylor/huma/issues/804
		Summary:     "Get a user's profile",
		Description: "Get a user's displayable profile",
	}, HandleGetUserProfile)
}

type InternalUser struct {
	RID         rid.RID   `json:"rid" readOnly:"true"`
	CreatedAt   time.Time `json:"createdAt" readOnly:"true"`
	Slug        *string   `json:"slug,omitempty"`
	DisplayName *string   `json:"displayName,omitempty"`
	BioText     *string   `json:"bioText,omitempty"`
	AvatarUrl   string    `json:"avatarUrl,omitempty" readOnly:"true"`
}

type ProfileUnlockedAchievements struct {
	DeveloperSlug string `json:"developerSlug" doc:"The developer associated with the achievement and game"`
	GameSlug      string `json:"gameSlug" doc:"The slug of the game that the unlocked achievement belongs to"`
	GameName      string `json:"gameName" doc:"The name of the game that the unlocked achievement belongs to"`
	Slug          string `json:"slug" doc:"The slug of the achievement that was unlocked"`
	Name          string `json:"name" doc:"The name of the achievement that was unlocked"`
	Description   string `json:"description" doc:"The description of the achievement that was unlocked"`
}

func (i *ProfileUnlockedAchievements) MapFromRow(row query.GetUserRecentAchievementsRow) {
	*i = ProfileUnlockedAchievements{
		DeveloperSlug: row.DeveloperSlug,
		GameSlug:      row.GameSlug,
		GameName:      row.GameName,
		Slug:          row.Slug,
		Name:          row.Name,
		Description:   row.Description,
	}
}

type ProfileOtherUserUnlockedAchievements struct {
	ProfileUnlockedAchievements
	UserRID          rid.RID
	UserFriendlyName string `json:"userFriendlyName" doc:"The best available name that can be displayed on screen for a human reader. Will be the user's display name if they have one, otherwise it will be their slug."`
}

func (i *ProfileOtherUserUnlockedAchievements) MapFromRow(row query.GetOtherUserRecentAchievementsRow) {
	*i = ProfileOtherUserUnlockedAchievements{
		ProfileUnlockedAchievements: ProfileUnlockedAchievements{
			DeveloperSlug: row.DeveloperSlug,
			GameSlug:      row.GameSlug,
			GameName:      row.GameName,
			Slug:          row.Slug,
			Name:          row.Name,
			Description:   row.Description,
		},
		UserRID: rid.RID{
			Prefix: auth.UserRidPrefix,
			ID:     row.UserUuid,
		},
		UserFriendlyName: row.UserFriendlyName,
	}
}

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

	return UserProfile{
		User: InternalUser{
			RID: rid.RID{
				Prefix: auth.UserRidPrefix,
				ID:     sessionProfile.Uuid,
			},
			CreatedAt:   sessionProfile.CreatedAt,
			Slug:        &sessionProfile.Slug,
			DisplayName: &sessionProfile.DisplayName,
		},
		UnlockedAchievements:  unlocks,
		OtherUserAchievements: otherUserUnlocks,
	}, nil
}

type GetSessionResponse struct {
	Body UserProfile
}

func HandleGetSessionProfile(ctx context.Context, input *struct{}) (*GetSessionResponse, error) {
	principal, hasPrincipal := auth.GetPrincipal(ctx)
	if !hasPrincipal {
		// shouldn't ever get here due to middleware check
		return nil, huma.Error401Unauthorized("no session")
	}

	userUuid := principal.User.Uuid

	profile, err := GetUserProfile(ctx, userUuid)
	if err != nil {
		return nil, err
	}

	return &GetSessionResponse{Body: profile}, nil
}

type PostSessionRequest struct {
	Body UserProfile
}

func HandlePostSessionProfile(ctx context.Context, input *PostSessionRequest) (*PostSessionRequest, error) {
	principal, hasPrincipal := auth.GetPrincipal(ctx)
	if !hasPrincipal {
		// shouldn't ever get here due to middleware check
		return nil, huma.Error401Unauthorized("no session")
	}

	updateErr := db.Queries.UpdateSessionProfile(ctx, query.UpdateSessionProfileParams{
		Uuid:           principal.User.Uuid,
		NewSlug:        input.Body.User.Slug,
		NewDisplayName: input.Body.User.DisplayName,
		// TODO: BioText
	})

	if errors.Is(updateErr, db.ErrSlugAlreadyInUse) {
		//return nil, &ConflictSignUpSlug{
		//	Location: "body.slug",
		//	Slug:     registerBody.Slug,
		//}
		return nil, huma.Error409Conflict("that slug is already in use")
	} else if updateErr != nil {
		return nil, updateErr
	}

	return nil, nil
}

const GameRidPrefix = "g"
const GameTokenRidPrefix = "gt"

type Developer struct {
	FriendlyName string `json:"friendlyName" readOnly:"true"`
}

type InternalGame struct {
	RID          rid.RID   `json:"rid"`
	Developer    Developer `json:"developer,omitempty" readOnly:"true"`
	FriendlyName string    `json:"friendlyName" readOnly:"true"`
}

type GameToken struct {
	RID       rid.RID      `json:"rid" readOnly:"true"`
	CreatedAt time.Time    `json:"createdAt" readOnly:"true"`
	ExpiresAt time.Time    `json:"expiresAt"`
	Comment   string       `json:"comment"`
	Game      InternalGame `json:"game"`
}

func (t *GameToken) MapFromRow(row query.FindUserGameTokensRow) {
	*t = GameToken{
		RID: rid.RID{
			Prefix: GameTokenRidPrefix,
			ID:     row.Uuid,
		},
		CreatedAt: row.CreatedAt,
		ExpiresAt: row.ExpiresAt,
		Comment:   row.Comment,
		Game: InternalGame{
			RID: rid.RID{
				Prefix: GameRidPrefix,
				ID:     row.GameUuid,
			},
			Developer: Developer{
				FriendlyName: row.DeveloperSlug,
			},
			FriendlyName: row.GameSlug,
		},
	}
}

type GameTokenList struct {
	Tokens []GameToken `json:"tokens"`
}

type GetSessionGameTokensResponse struct {
	Body GameTokenList
}

func HandleGetSessionGameTokens(ctx context.Context, input *struct{}) (*GetSessionGameTokensResponse, error) {
	principal, hasPrincipal := auth.GetPrincipal(ctx)
	if !hasPrincipal {
		// shouldn't ever get here due to middleware check
		return nil, huma.Error401Unauthorized("no session")
	}

	foundTokens, err := db.Queries.FindUserGameTokens(ctx, principal.User.Uuid)
	if err != nil {
		return nil, err
	}

	gameTokens := make([]GameToken, len(foundTokens))
	for tokenIdx, _ := range foundTokens {
		gameTokens[tokenIdx].MapFromRow(foundTokens[tokenIdx])
	}

	return &GetSessionGameTokensResponse{Body: GameTokenList{Tokens: gameTokens}}, nil
}

type PostSessionGameTokenRequest struct {
	Body GameToken
}

type PostSessionGameTokenResponse struct {
	Body GameToken
}

func HandlePostSessionGameToken(ctx context.Context, input *PostSessionGameTokenRequest) (*PostSessionGameTokenResponse, error) {
	principal, hasPrincipal := auth.GetPrincipal(ctx)
	if !hasPrincipal {
		// shouldn't ever get here due to middleware check
		return nil, huma.Error401Unauthorized("no session")
	}

	// TODO: is there a way to add a custom validation in huma for RID prefix?
	if input.Body.Game.RID.Prefix != GameRidPrefix {
		return nil, huma.Error400BadRequest("invalid game id")
	}

	createdToken, createErr := db.Queries.CreateGameToken(ctx, query.CreateGameTokenParams{
		ExpiresAt: input.Body.ExpiresAt,
		Comment:   input.Body.Comment,
		UserUuid:  principal.User.Uuid,
		GameUuid:  input.Body.Game.RID.ID,
	})
	if createErr != nil {
		return nil, createErr
	}

	return &PostSessionGameTokenResponse{
		Body: GameToken{
			RID:       rid.From(GameTokenRidPrefix, createdToken.Uuid),
			CreatedAt: createdToken.CreatedAt,
			ExpiresAt: createdToken.ExpiresAt,
			Comment:   createdToken.Comment,
			Game: InternalGame{
				RID: rid.From(GameRidPrefix, createdToken.GameUuid),
				Developer: Developer{
					FriendlyName: createdToken.DeveloperSlug,
				},
				FriendlyName: createdToken.GameSlug,
			},
		},
	}, nil
}

type DeleteSessionGameTokenResponse struct {
	GameTokenRID rid.RID `query:"tokenRID" example:"gt_31F0otb4FIVRqQWdsISFl"`
}

func HandleDeleteSessionGameToken(ctx context.Context, input *DeleteSessionGameTokenResponse) (*struct{}, error) {
	principal, hasPrincipal := auth.GetPrincipal(ctx)
	if !hasPrincipal {
		// shouldn't ever get here due to middleware check
		return nil, huma.Error401Unauthorized("no session")
	}

	// TODO: is there a way to add a custom validation in huma for RID prefix?
	if input.GameTokenRID.Prefix != GameTokenRidPrefix {
		return nil, huma.Error400BadRequest("invalid game id")
	}

	rows, err := db.Queries.ExpireToken(ctx, query.ExpireTokenParams{
		UserUuid: principal.User.Uuid,
		Uuid:     input.GameTokenRID.ID,
	})
	if err != nil {
		return nil, err
	}

	if rows == 0 {
		return nil, huma.Error404NotFound("game token not found")
	}

	return &struct{}{}, nil
}

type SearchUsersRequest struct {
	SlugLike string                       `query:"slugLike" required:"true"`
	After    validation.Optional[rid.RID] `query:"after,omitempty"`
	Limit    validation.Optional[int]     `query:"limit" minimum:"10" maximum:"50" doc:"default = 10"`
}

type InternalUserList struct {
	Users []InternalUser `json:"users"`
}

type SearchUsersResponse struct {
	Body InternalUserList
}

func HandleSearchUsers(ctx context.Context, input *SearchUsersRequest) (*SearchUsersResponse, error) {
	// TODO: a huma validator for rid prefix...
	if input.After.HasValue && input.After.Value.Prefix != auth.UserRidPrefix {
		return nil, huma.Error400BadRequest("invalid user id")
	}

	principal, hasPrincipal := auth.GetPrincipal(ctx)
	isAdmin := hasPrincipal && auth.IsAdmin(principal.User)

	var principalUuid uuid.UUID
	if hasPrincipal {
		principalUuid = principal.User.Uuid
	}

	builder := db.DB.Builder().
		Select("u.uuid", "u.created_at", "u.slug", "coalesce(uldn.display_name, '')").
		From("users u").
		JoinClause("left outer join user_latest_display_name uldn on u.id = uldn.user_id").
		JoinClause("left outer join user_latest_email ule on u.id = ule.user_id and (? or (? and u.uuid = ?))", isAdmin, hasPrincipal, principalUuid).
		Where("u.slug like ?", "%"+input.SlugLike+"%").
		OrderBy("u.uuid desc")

	if input.After.HasValue {
		builder = builder.Where("u.uuid > ?", input.After.Value.ID)
	}

	limit := input.Limit.ValueOr(10)
	builder = builder.Limit(uint64(limit))

	rows, queryErr := db.DB.Query(ctx, builder)
	if queryErr != nil && !errors.Is(queryErr, sql.ErrNoRows) {
		return nil, eris.Wrap(queryErr, "")
	}

	defer rows.Close()

	var items []InternalUser
	for rows.Next() {
		var userUuid uuid.UUID
		var item InternalUser

		if scanErr := rows.Scan(
			&userUuid,
			&item.CreatedAt,
			&item.Slug,
			&item.DisplayName,
			// TODO: BioText, AvatarUrl
		); scanErr != nil {
			return nil, eris.Wrap(scanErr, "")
		}

		item.RID = rid.From(auth.UserRidPrefix, userUuid)
		items = append(items, item)
	}

	return &SearchUsersResponse{
		Body: InternalUserList{
			Users: items,
		},
	}, nil
}

type GetUserProfileRequest struct {
	UserRID rid.RID `query:"userRid" required:"true"`
}

type GetUserProfileResponse struct {
	Body UserProfile
}

func HandleGetUserProfile(ctx context.Context, input *GetUserProfileRequest) (*GetUserProfileResponse, error) {
	// TODO: huma validator for rid prefix...
	if input.UserRID.Prefix != auth.UserRidPrefix {
		return nil, huma.Error400BadRequest("invalid user id")
	}

	profile, err := GetUserProfile(ctx, input.UserRID.ID)
	if err != nil {
		return nil, err
	}

	return &GetUserProfileResponse{Body: profile}, nil
}

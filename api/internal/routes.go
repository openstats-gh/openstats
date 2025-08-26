package internal

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"image/png"
	"net/http"
	"strconv"
	"time"

	"github.com/buckket/go-blurhash"
	"github.com/danielgtaylor/huma/v2"
	"github.com/dresswithpockets/openstats/app/auth"
	"github.com/dresswithpockets/openstats/app/db"
	"github.com/dresswithpockets/openstats/app/db/query"
	"github.com/dresswithpockets/openstats/app/media"
	"github.com/dresswithpockets/openstats/app/rid"
	"github.com/dresswithpockets/openstats/app/users"
	"github.com/dresswithpockets/openstats/app/validation"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
)

func RegisterRoutes(api huma.API) {
	internalApi := huma.NewGroup(api, "/internal")

	var sessionCookieSecurityMap = []map[string][]string{{"SessionCookie": {}}}
	var requireUserSessionMiddlewares = huma.Middlewares{
		auth.UserAuthHandler,
		auth.CreateRequireUserAuthHandler(internalApi),
	}
	var disallowUserSessionMiddlewares = huma.Middlewares{
		auth.UserAuthHandler,
		auth.CreateRequireNoUserAuthHandler(internalApi),
	}

	huma.Register(internalApi, huma.Operation{
		Method:      http.MethodGet,
		Path:        "/send-slug-reminder",
		OperationID: "send-slug-reminder",
		Summary:     "Send slug reminder",
		Description: "Send an email to the email provided containing a list of all users associated with the email",
		Errors:      []int{http.StatusUnauthorized, http.StatusBadRequest},
		Tags:        []string{"Internal"},

		Middlewares: disallowUserSessionMiddlewares,
	}, HandleSendSlugReminder)

	sessionApi := huma.NewGroup(internalApi, "/session")
	sessionApi.UseSimpleModifier(func(op *huma.Operation) {
		op.Tags = append(op.Tags, "Internal/Session")
	})

	huma.Register(sessionApi, huma.Operation{
		Method:      http.MethodPost,
		Path:        "/sign-up",
		OperationID: "sign-up",
		Errors: []int{
			http.StatusUnauthorized,
			http.StatusConflict,
		},
		Summary:     "Sign up",
		Description: "Create a new user and sign into a new session as the new user",

		Middlewares: disallowUserSessionMiddlewares,
	}, HandlePostSignUp)

	huma.Register(sessionApi, huma.Operation{
		Method:      http.MethodPost,
		Path:        "/sign-in",
		OperationID: "sign-in",
		Summary:     "Sign in",
		Description: "Sign into a new session as an existing user",
		Errors:      []int{http.StatusUnauthorized},

		Middlewares: disallowUserSessionMiddlewares,
	}, HandlePostSignIn)

	huma.Register(sessionApi, huma.Operation{
		Method:      http.MethodPost,
		Path:        "/sign-out",
		OperationID: "sign-out",
		Summary:     "Sign out",
		Description: "Sign out of the current session, and invalidate the session token",

		Security:    sessionCookieSecurityMap,
		Middlewares: requireUserSessionMiddlewares,
	}, HandlePostSignOut)

	huma.Register(sessionApi, huma.Operation{
		Method:      http.MethodGet,
		Path:        "/",
		OperationID: "get-session",
		Summary:     "Get session summary",
		Description: "Get details about the current authenticated session and the associated user",
		Errors:      []int{http.StatusUnauthorized},

		Middlewares: requireUserSessionMiddlewares,
	}, HandleGetSession)

	huma.Register(sessionApi, huma.Operation{
		Method:      http.MethodPost,
		Path:        "/add-email",
		OperationID: "add-email",
		Summary:     "Add an email",
		Description: "Sends a confirmation to the email; once confirmed by /confirm-email, the email will be associated with the current session's user",
		Errors:      []int{http.StatusUnauthorized, http.StatusConflict},

		Middlewares: requireUserSessionMiddlewares,
	}, HandleAddEmail)

	huma.Register(sessionApi, huma.Operation{
		Method:      http.MethodPost,
		Path:        "/confirm-email",
		OperationID: "confirm-email",
		Summary:     "Confirm an email",
		Description: "Validates an email confirmation TOTP; if successful, the email will be marked as verified",
		Errors:      []int{http.StatusUnauthorized},

		Middlewares: requireUserSessionMiddlewares,
	}, HandleConfirmEmail)

	huma.Register(sessionApi, huma.Operation{
		Method:      http.MethodPost,
		Path:        "/remove-email",
		OperationID: "remove-email",
		Summary:     "Remove an email",
		Description: "Removes one of the emails from the current session's user",
		Errors:      []int{http.StatusUnauthorized},

		Middlewares: requireUserSessionMiddlewares,
	}, HandleRemoveEmail)

	huma.Register(sessionApi, huma.Operation{
		Method:      http.MethodPost,
		Path:        "/replace-password",
		OperationID: "replace-password",
		Summary:     "Change user's password",
		Description: "Changes the current session user's password",
		Errors:      []int{http.StatusUnauthorized},

		Middlewares: requireUserSessionMiddlewares,
	}, HandleChangePassword)

	huma.Register(sessionApi, huma.Operation{
		Method:      http.MethodGet,
		Path:        "/profile",
		OperationID: "get-session-profile",
		Summary:     "Get user's profile",
		Description: "Get profile of current authenticated user",

		Middlewares: requireUserSessionMiddlewares,
	}, HandleGetSessionProfile)

	huma.Register(sessionApi, huma.Operation{
		Method:      http.MethodPost,
		Path:        "/profile",
		OperationID: "update-session-profile",
		Summary:     "Update user's profile",
		Description: "Update profile of current authenticated user",

		Middlewares: requireUserSessionMiddlewares,
	}, HandlePostSessionProfile)

	huma.Register(sessionApi, huma.Operation{
		Method:      http.MethodPost,
		Path:        "/profile/avatar",
		OperationID: "update-session-avatar",
		Summary:     "Update user's avatar",
		Description: "Update avatar of current authenticated user",

		Middlewares: requireUserSessionMiddlewares,
	}, HandlePostSessionAvatar)

	huma.Register(sessionApi, huma.Operation{
		Method:      http.MethodGet,
		Path:        "/tokens",
		OperationID: "get-game-tokens",
		Summary:     "Get user's tokens",
		Description: "Get all of the current user's tokens",

		Middlewares: requireUserSessionMiddlewares,
	}, HandleGetSessionGameTokens)

	huma.Register(sessionApi, huma.Operation{
		Method:      http.MethodPost,
		Path:        "/tokens",
		OperationID: "create-game-token",
		Summary:     "Create a new token",
		Description: "Create a new token for the current user",
		Errors:      []int{http.StatusBadRequest},

		Middlewares: requireUserSessionMiddlewares,
	}, HandlePostSessionGameToken)

	huma.Register(sessionApi, huma.Operation{
		Method:      http.MethodDelete,
		Path:        "/tokens/{tokenRID}",
		OperationID: "delete-game-token",
		Summary:     "Invalidate a token",
		Description: "Invalidate one of the current user's tokens",
		Errors:      []int{http.StatusBadRequest},

		Middlewares: requireUserSessionMiddlewares,
	}, HandleDeleteSessionGameToken)

	userApi := huma.NewGroup(internalApi, "/users/v1")
	userApi.UseSimpleModifier(func(op *huma.Operation) {
		op.Tags = append(op.Tags, "Internal/Users")
	})
	huma.Register(userApi, huma.Operation{
		Method:      http.MethodGet,
		Path:        "/",
		OperationID: "search-users",
		Summary:     "Search users",
		Description: "Search all users by various criteria",

		Middlewares: requireUserSessionMiddlewares,
	}, HandleSearchUsers)

	huma.Register(userApi, huma.Operation{
		Method:      http.MethodGet,
		Path:        "/{user}/profile",
		OperationID: "get-user-profile",
		Summary:     "Get a user's profile",
		Description: "Get a user's displayable profile",

		Middlewares: requireUserSessionMiddlewares,
	}, HandleGetUserProfile)
}

type SendEmailConfInput struct {
	Body struct {
		Email string `json:"email" format:"email"`
	}
}

type SendEmailConfOutput struct{}

func HandleAddEmail(ctx context.Context, input *SendEmailConfInput) (output *SendEmailConfOutput, err error) {
	principal, hasPrincipal := auth.GetPrincipal(ctx)
	if !hasPrincipal {
		// shouldn't ever get here due to middleware check
		return nil, huma.Error401Unauthorized("no session")
	}

	var userEmail query.UserEmail
	userEmail, err = db.Queries.AddOrGetUserEmail(ctx, query.AddOrGetUserEmailParams{
		UserID: principal.User.ID,
		Email:  input.Body.Email,
	})
	if err != nil {
		return
	}

	if userEmail.ConfirmedAt.Valid || userEmail.Email != input.Body.Email {
		return nil, huma.Error409Conflict("email already associated with this user")
	}

	var hmacSecret string
	hmacSecret, err = db.Queries.SecretRead(ctx, query.SecretReadParams{
		Path: db.PrivateUser2faHmacSecretPath,
		Key:  strconv.FormatInt(int64(principal.User.ID), 10),
	})

	if err != nil {
		return nil, eris.Wrap(err, "there was an error creating your 2FA TOTP code")
	}

	err = SendEmailConfirmation(ctx, hmacSecret, userEmail.Email)
	if err != nil {
		return nil, eris.Wrap(err, "there was an error sending your 2FA TOTP code")
	}

	return &SendEmailConfOutput{}, nil
}

type ConfirmEmailInput struct {
	Body struct {
		Email string `json:"email" format:"email"`
		Code  string `json:"code"`
	}
}

type EmailValidationResult struct {
	Validated bool `json:"validated"`
}

type ConfirmEmailOutput struct {
	Body EmailValidationResult
}

func HandleConfirmEmail(ctx context.Context, input *ConfirmEmailInput) (output *ConfirmEmailOutput, err error) {
	principal, hasPrincipal := auth.GetPrincipal(ctx)
	if !hasPrincipal {
		// shouldn't ever get here due to middleware check
		return nil, huma.Error401Unauthorized("no session")
	}

	validated, validateErr := ValidateUserEmail(ctx, principal.User.ID, input.Body.Email, input.Body.Code)
	if validateErr != nil {
		return nil, validateErr
	}

	return &ConfirmEmailOutput{Body: EmailValidationResult{Validated: validated}}, nil
}

type RemoveEmailInput struct {
	Body struct {
		Email string `json:"email" format:"email"`
	}
}
type RemoveEmailOutput struct{}

func HandleRemoveEmail(ctx context.Context, input *RemoveEmailInput) (output *RemoveEmailOutput, err error) {
	principal, hasPrincipal := auth.GetPrincipal(ctx)
	if !hasPrincipal {
		return nil, huma.Error401Unauthorized("no session")
	}

	_, err = db.Queries.RemoveEmail(ctx, query.RemoveEmailParams{
		UserID: principal.User.ID,
		Email:  input.Body.Email,
	})

	if eris.Is(err, sql.ErrNoRows) {
		return nil, huma.Error404NotFound("that email isn't associated with this user")
	}

	return &RemoveEmailOutput{}, err
}

type ChangePasswordInput struct {
	Body struct {
		CurrentPassword string `json:"currentPassword" required:"true" pattern:"[a-zA-Z0-9!@#$%^&*]+" patternDescription:"alphanum with specials" minLength:"10" maxLength:"32"`
		NewPassword     string `json:"newPassword" required:"true" pattern:"[a-zA-Z0-9!@#$%^&*]+" patternDescription:"alphanum with specials" minLength:"10" maxLength:"32"`
	}
}
type ChangePasswordOutput struct{}

func HandleChangePassword(ctx context.Context, input *ChangePasswordInput) (output *ChangePasswordOutput, err error) {
	principal, hasPrincipal := auth.GetPrincipal(ctx)
	if !hasPrincipal {
		return nil, huma.Error401Unauthorized("no session")
	}

	err = auth.ReplaceUserPassword(ctx, principal.User.ID, input.Body.CurrentPassword, input.Body.NewPassword)
	if err != nil {
		return nil, err
	}

	return &ChangePasswordOutput{}, nil
}

type InternalUser struct {
	RID         rid.RID       `json:"rid" readOnly:"true"`
	CreatedAt   time.Time     `json:"createdAt" readOnly:"true"`
	Slug        *string       `json:"slug,omitempty"`
	DisplayName *string       `json:"displayName,omitempty"`
	BioText     *string       `json:"bioText,omitempty"`
	Avatar      *users.Avatar `json:"avatar,omitempty" readOnly:"true"`
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

	if db.IsUniqueConstraintErr(updateErr) {
		//return nil, &ConflictSignUpSlug{
		//	Location: "body.slug",
		//	Slug:     registerBody.Slug,
		//}
		return nil, huma.Error409Conflict("that slug is already in use")
	}

	return nil, updateErr
}

type PostAvatarInput struct {
	Body []byte
}

type PostAvatarOutput struct {
	Location string `header:"Location"`
}

func HandlePostSessionAvatar(ctx context.Context, input *PostAvatarInput) (*PostAvatarOutput, error) {
	principal, hasPrincipal := auth.GetPrincipal(ctx)
	if !hasPrincipal {
		return nil, huma.Error401Unauthorized("no session")
	}

	decodedPng, pngErr := png.Decode(bytes.NewBuffer(input.Body))
	if pngErr != nil {
		return nil, pngErr
	}

	blur, blurErr := blurhash.Encode(4, 4, decodedPng)
	if blurErr != nil {
		return nil, blurErr
	}

	var newAvatar query.UserAvatar
	transactErr := db.DB.Transact(ctx, func(c context.Context, queries *query.Queries) (err error) {
		newAvatar, err = db.Queries.AddUserAvatar(ctx, query.AddUserAvatarParams{
			Blurhash: blur,
			UserUuid: principal.User.Uuid,
		})
		if err != nil {
			return err
		}

		return media.WriteAvatar(input.Body, "users", newAvatar.Uuid)
	})

	if transactErr != nil {
		return nil, transactErr
	}

	return &PostAvatarOutput{
		Location: media.GetAvatarUrl("users", newAvatar.Uuid),
	}, nil
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
		Select("u.uuid", "u.created_at", "u.slug", "coalesce(uldn.display_name, ''), ua.uuid as avatar_uuid, ua.blurhash as avatar_blurhash").
		From("users u").
		JoinClause("left outer join user_latest_display_name uldn on u.id = uldn.user_id").
		JoinClause("left outer join user_latest_email ule on u.id = ule.user_id and (? or (? and u.uuid = ?))", isAdmin, hasPrincipal, principalUuid).
		JoinClause("left outer join user_avatar ua on u.id = ua.user_id").
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

		var avatarUuid uuid.NullUUID
		var avatarBlurhash *string

		if scanErr := rows.Scan(
			&userUuid,
			&item.CreatedAt,
			&item.Slug,
			&item.DisplayName,
			&avatarUuid,
			&avatarBlurhash,
			// TODO: BioText
		); scanErr != nil {
			return nil, eris.Wrap(scanErr, "")
		}

		item.RID = rid.From(auth.UserRidPrefix, userUuid)
		if avatarUuid.Valid && avatarBlurhash != nil {
			item.Avatar = &users.Avatar{
				Url:      media.GetAvatarUrl("users", avatarUuid.UUID),
				Blurhash: *avatarBlurhash,
			}
		}
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

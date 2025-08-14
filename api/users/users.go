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
	"github.com/rotisserie/eris"
	"math/rand"
	"net/http"
	"time"
)

// User resource returned by users/ endpoints
type User struct {
	RID         rid.RID              `json:"rid" readOnly:"true"`
	CreatedAt   validation.EpochTime `json:"createdAt" readOnly:"true"`
	Slug        string               `json:"slug"`
	DisplayName string               `json:"displayName,omitempty"`
	BioText     string               `json:"bioText,omitempty"`
	AvatarUrl   string               `json:"avatarUrl,omitempty" readOnly:"true"`
	Email       string               `json:"email,omitempty"`
	Password    string               `json:"password,omitempty" writeOnly:"true"`
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

func RegisterRoutes(api huma.API) {
	usersApi := huma.NewGroup(api, "/users/v1")

	requireGameTokenAuthHandler := auth.CreateRequireGameTokenAuthHandler(usersApi)
	requireGameSessionAuthHandler := auth.CreateRequireGameSessionAuthHandler(usersApi)
	usersApi.UseSimpleModifier(func(op *huma.Operation) {
		op.Tags = append(op.Tags, "Users")
	})

	huma.Register(usersApi, huma.Operation{
		Path:        "/",
		OperationID: "internal-i-users-get",
		Method:      http.MethodGet,
		Security:    []map[string][]string{{"GameToken": {}}},
		Middlewares: huma.Middlewares{auth.GameTokenAuthHandler, requireGameTokenAuthHandler}, // TODO: https://github.com/danielgtaylor/huma/issues/804
		Summary:     "Get users",
		Description: "Search all users by various criteria",
	}, HandleSearchUsers)

	huma.Register(usersApi, huma.Operation{
		Path:        "/{user}",
		OperationID: "get-user",
		Method:      http.MethodGet,
		Security:    []map[string][]string{{"GameToken": {}}},
		Middlewares: huma.Middlewares{auth.GameTokenAuthHandler, requireGameTokenAuthHandler},
		Summary:     "Get user",
		Description: "Get a user by RID, or get the user associated with the Game Token if @me is provided instead of an RID",
	}, HandleGetUser)

	huma.Register(usersApi, huma.Operation{
		Path:        "/{user}/games/{game}/sessions",
		OperationID: "users-create-game-session",
		Method:      http.MethodPost,
		Security:    []map[string][]string{{"GameToken": {}}},
		Errors:      []int{http.StatusBadRequest, http.StatusUnauthorized},
		Middlewares: huma.Middlewares{auth.GameTokenAuthHandler, requireGameTokenAuthHandler}, // TODO: https://github.com/danielgtaylor/huma/issues/804
		Summary:     "Create a game session",
		Description: "Create a new game session. Game Sessions are used to track playtime, stats, and achievements.",
	}, HandleCreateGameSession)

	huma.Register(usersApi, huma.Operation{
		Path:        "/{user}/games/{game}/sessions/{session}/heartbeat",
		OperationID: "users-game-session-heartbeat",
		Method:      http.MethodPost,
		Security:    []map[string][]string{{"GameSession": {}}},
		Middlewares: huma.Middlewares{auth.GameSessionAuthHandler, requireGameSessionAuthHandler},
		Summary:     "Refresh the game session",
		Description: "Refresh the game session. Update the session's last pulse, and generate a new game session token if the expiration is too close.",
	}, HandleHeartbeatGameSession)

	huma.Register(usersApi, huma.Operation{
		Path:        "/{user}/games/{game}/achievements",
		OperationID: "internal-j-users-get-achievements",
		Method:      http.MethodGet,
		Security:    []map[string][]string{{"GameSession": {}}},
		Middlewares: huma.Middlewares{auth.GameSessionAuthHandler, requireGameSessionAuthHandler},
		Summary:     "Get a user's achievements",
		Description: "Get a user's achievement progress for the game associated with the session",
	}, HandleGetUserAchievements)

	huma.Register(usersApi, huma.Operation{
		Path:        "/{user}/games/{game}/achievements",
		OperationID: "users-game-session-set-progress",
		Method:      http.MethodPost,
		Security:    []map[string][]string{{"GameSession": {}}},
		Middlewares: huma.Middlewares{auth.GameSessionAuthHandler, requireGameSessionAuthHandler},
		Summary:     "Add achievement progress",
		Description: "Add new progress to one or multiple achievements for a particular user. Any progress that's lower than the user's current progress for the associated achievement will be ignored.",
	}, HandleSetUserProgress)
}

type SearchUsersRequest struct {
	SlugLike string                       `query:"slugLike" required:"true"`
	After    validation.Optional[rid.RID] `query:"after,omitempty"`
	Limit    validation.Optional[int]     `query:"limit" minimum:"10" maximum:"50" doc:"default = 10"`
}

type UserList struct {
	Users []User `json:"users"`
}

type SearchUsersResponse struct {
	Body UserList
}

func HandleSearchUsers(ctx context.Context, input *SearchUsersRequest) (*SearchUsersResponse, error) {
	// TODO: a huma validator for rid prefix...
	if input.After.HasValue && input.After.Value.Prefix != auth.UserRidPrefix {
		return nil, huma.Error400BadRequest("invalid user id")
	}

	// TODO: searching with the GameToken principal should only yield results for players which have an association with
	//       the game (maybe not?)

	builder := db.DB.Builder().
		Select("u.uuid", "u.created_at", "u.slug", "coalesce(uldn.display_name, '')").
		From("users u").
		JoinClause("left outer join user_latest_display_name uldn on u.id = uldn.user_id").
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

	var items []User
	for rows.Next() {
		var userUuid uuid.UUID
		var item User

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
		Body: UserList{
			Users: items,
		},
	}, nil
}

type GetUserInput struct {
	User string `path:"user"`
}

type GetUserOutput struct {
	Body User
}

func HandleGetUser(ctx context.Context, input *GetUserInput) (*GetUserOutput, error) {
	principal, hasPrincipal := auth.GetGameTokenPrincipal(ctx)
	if !hasPrincipal {
		return nil, huma.Error401Unauthorized("Game Token required to get users")
	}

	userUuid := principal.UserRid.ID
	if input.User != "@me" {
		userRid, ridErr := rid.ParseString(input.User)
		if ridErr != nil {
			return nil, ridErr
		}

		userUuid = userRid.ID
	}

	user, err := db.Queries.GetUserWithName(ctx, userUuid)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, huma.Error404NotFound("User not found")
	}

	if err != nil {
		return nil, err
	}

	return &GetUserOutput{
		Body: User{
			RID:         rid.From(auth.UserRidPrefix, userUuid),
			CreatedAt:   validation.ToEpochTime(user.CreatedAt),
			Slug:        user.Slug,
			DisplayName: user.DisplayName,
		},
	}, nil
}

type Game struct {
	RID       rid.RID              `json:"rid" readOnly:"true"`
	CreatedAt validation.EpochTime `json:"createdAt" readOnly:"true"`
	Slug      string               `json:"slug"`
}

type GameSession struct {
	RID            rid.RID              `json:"rid" readOnly:"true"`
	LastPulse      validation.EpochTime `json:"lastPulse" readOnly:"true"`
	NextPulseAfter int                  `json:"nextPulseAfter" readOnly:"true" doc:"the number of seconds before your next heartbeat. Always send your next heartbeat close to this amount of time after the response is received, otherwise the API cannot guarantee that the JWT will continue to be valid for the next heartbeat."`
	User           User                 `json:"user" readOnly:"true"`
	Game           Game                 `json:"game" readOnly:"true"`
}

type CreateGameSessionInput struct {
	User rid.RID `path:"user"`
	Game rid.RID `path:"game"`
}

type CreateGameSessionOutput struct {
	Token string "header:\"X-Game-Session-Token\" doc:\"Contains the JWT authorized for the new Game Session. Authenticate your future requests in the `Authorization` header as a Bearer token.\""
	Body  GameSession
}

func HandleCreateGameSession(ctx context.Context, input *CreateGameSessionInput) (*CreateGameSessionOutput, error) {
	principal, hasPrincipal := auth.GetGameTokenPrincipal(ctx)
	if !hasPrincipal {
		return nil, huma.Error401Unauthorized("creating a Game Session requires a user-supplied Game Token")
	}

	if input.User.ID != principal.UserRid.ID || input.Game.ID != principal.GameRid.ID {
		return nil, huma.Error401Unauthorized("sessions may only be created for the user and game that the Game Token is associated with")
	}

	signedToken, gameSession, err := auth.CreateGameSessionToken(ctx, principal.TokenUuid, principal.UserRid, principal.GameRid)
	if err != nil {
		return nil, err
	}

	user, userErr := db.Queries.FindUserById(ctx, gameSession.UserID)
	if userErr != nil {
		return nil, userErr
	}

	game, gameErr := db.Queries.FindGameById(ctx, gameSession.GameID)
	if gameErr != nil {
		return nil, gameErr
	}

	// we jitter 360 +/- 180 seconds
	pulseJitter := rand.Intn(360) - 180
	nextPulseDuration := 360 + pulseJitter

	return &CreateGameSessionOutput{
		Token: signedToken,
		Body: GameSession{
			RID:            rid.From(auth.GameSessionRidPrefix, gameSession.Uuid),
			LastPulse:      validation.ToEpochTime(gameSession.LastPulseAt),
			NextPulseAfter: nextPulseDuration,
			User: User{
				RID:       rid.From(auth.UserRidPrefix, user.Uuid),
				CreatedAt: validation.ToEpochTime(user.CreatedAt),
				Slug:      user.Slug,
			},
			Game: Game{
				RID:       rid.From(auth.GameRidPrefix, game.Uuid),
				CreatedAt: validation.ToEpochTime(game.CreatedAt),
				Slug:      game.Slug,
			},
		},
	}, nil
}

type HeartbeatGameSessionInput struct {
	User    rid.RID `path:"user"`
	Game    rid.RID `path:"game"`
	Session rid.RID `path:"session"`
}

type HeartbeatGameSessionOutput struct {
	Token *string "header:\"X-Game-Session-Token\" doc:\"If null, continue to use the token that this request was authenticated with. Otherwise, contains a new JWT that you should authenticate future requests with.\""
	Body  GameSession
}

func HandleHeartbeatGameSession(ctx context.Context, input *HeartbeatGameSessionInput) (output *HeartbeatGameSessionOutput, err error) {
	principal, hasPrincipal := auth.GetGameSessionPrincipal(ctx)
	if !hasPrincipal || input.User.ID != principal.UserRid.ID || input.Game.ID != principal.GameRid.ID || input.Session.ID != principal.SessionRid.ID {
		return nil, huma.Error401Unauthorized("heartbeats must be authenticated using a Game Session Token")
	}

	// we jitter 360 +/- 180 seconds
	pulseJitter := rand.Intn(360) - 180
	nextPulseDuration := 360 + pulseJitter

	// we subtract 1 minute from the nextPulseDirection when getting the nextPulseTime as additional jitter for this
	// request. We really don't want the caller to have an expired token by the time this request makes its way to us -
	// otherwise they'd have to create a new game session!
	nextPulseTime := time.Now().UTC().Add(time.Duration(nextPulseDuration-60) * time.Second)

	var resultToken *string
	lastPulseAt := principal.LastPulse
	gameSessionRid := principal.SessionRid
	if nextPulseTime.After(principal.ExpiresAt) {
		// the next pulse will happen too close to the expiration, so we create a new token
		signedToken, gameSession, createErr := auth.CreateGameSessionToken(ctx, principal.GameTokenUuid, principal.UserRid, principal.GameRid)
		if createErr != nil {
			return nil, createErr
		}

		resultToken = &signedToken
		lastPulseAt = gameSession.LastPulseAt
		gameSessionRid = rid.From(auth.GameSessionRidPrefix, gameSession.Uuid)
	} else {
		lastPulseAt, err = db.Queries.HeartbeatGameSession(ctx, gameSessionRid.ID)
		if err != nil {
			return
		}
	}

	user, userErr := db.Queries.FindUser(ctx, principal.UserRid.ID)
	if userErr != nil {
		return nil, userErr
	}

	game, gameErr := db.Queries.FindGame(ctx, principal.GameRid.ID)
	if gameErr != nil {
		return nil, gameErr
	}

	output = &HeartbeatGameSessionOutput{
		Token: resultToken,
		Body: GameSession{
			RID:            gameSessionRid,
			LastPulse:      validation.ToEpochTime(lastPulseAt),
			NextPulseAfter: nextPulseDuration,
			User: User{
				RID:       principal.UserRid,
				CreatedAt: validation.ToEpochTime(user.CreatedAt),
				Slug:      user.Slug,
			},
			Game: Game{
				RID:       principal.GameRid,
				CreatedAt: validation.ToEpochTime(game.CreatedAt),
				Slug:      game.Slug,
			},
		},
	}

	return
}

type GetUserAchievementsRequest struct {
	User rid.RID `path:"user"`
	Game rid.RID `path:"game"`
}

type UserProgress struct {
	Progress map[string]int32 `json:"progress" doc:"a map of slugs to the user's current progress in the associated achievement'"`
}

type GetUserAchievementsResponse struct {
	Body UserProgress
}

func HandleGetUserAchievements(ctx context.Context, input *GetUserAchievementsRequest) (*GetUserAchievementsResponse, error) {
	principal, hasPrincipal := auth.GetGameSessionPrincipal(ctx)
	if !hasPrincipal {
		return nil, huma.Error401Unauthorized("invalid session")
	}

	if input.User.ID != principal.UserRid.ID {
		return nil, huma.Error401Unauthorized("you may only get progress for the same user that the session was created for")
	}

	if input.Game.ID != principal.GameRid.ID {
		return nil, huma.Error401Unauthorized("you may only get progress for the same game that the session was created for")
	}

	progressRows, dbErr := db.Queries.GetGameSessionUserProgress(ctx, query.GetGameSessionUserProgressParams{
		UserUuid: principal.UserRid.ID,
		GameUuid: principal.GameRid.ID,
	})
	if dbErr != nil {
		return nil, dbErr
	}

	resultMap := map[string]int32{}
	for _, progressRow := range progressRows {
		resultMap[progressRow.Slug] = progressRow.Progress
	}

	return &GetUserAchievementsResponse{
		Body: UserProgress{Progress: resultMap},
	}, nil
}

type SetUserProgressRequest struct {
	User rid.RID `path:"user"`
	Game rid.RID `path:"game"`
	Body UserProgress
}

type SetUserProgressResponse struct {
	Body UserProgress
}

func HandleSetUserProgress(ctx context.Context, input *SetUserProgressRequest) (*SetUserProgressResponse, error) {
	principal, hasPrincipal := auth.GetGameSessionPrincipal(ctx)
	if !hasPrincipal {
		return nil, huma.Error401Unauthorized("invalid session")
	}

	if input.User.ID != principal.UserRid.ID {
		return nil, huma.Error401Unauthorized("you may only update progress for the same user that the session was created for")
	}

	if input.Game.ID != principal.GameRid.ID {
		return nil, huma.Error401Unauthorized("you may only update progress for the same game that the session was created for")
	}

	var params []query.UpdateGameSessionUserProgressParams
	for slug, progress := range input.Body.Progress {
		params = append(params, query.UpdateGameSessionUserProgressParams{
			NewProgress:     progress,
			UserUuid:        input.User.ID,
			AchievementSlug: slug,
			GameUuid:        input.Game.ID,
		})
	}

	results := map[string]int32{}
	var batchErr error
	batchResults := db.Queries.UpdateGameSessionUserProgress(ctx, params)
	batchResults.QueryRow(func(i int, row query.UpdateGameSessionUserProgressRow, err error) {
		if errors.Is(err, sql.ErrNoRows) {
			return
		}

		if err != nil {
			batchErr = err
			return
		}

		results[row.Slug] = row.Progress
	})

	if batchErr != nil {
		return nil, batchErr
	}

	return &SetUserProgressResponse{
		Body: UserProgress{Progress: results},
	}, nil
}

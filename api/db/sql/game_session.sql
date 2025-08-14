-- name: GetGameSessionRidCounts :one
with target_user as (
    select count() as user_count from users where users.uuid = @user_uuid
), target_session as (
    select last_pulse_at
    from game_session
    where game_session.uuid = @session_uuid

), target_game as (
    select count() as game_count from game where game.uuid = @game_uuid
), disallow_jwt as (
    select count() as disallow_count from token_disallow_list tdl where tdl.token_id = @token_uuid
)
select target_user.user_count, target_session.last_pulse_at, target_game.game_count, disallow_jwt.disallow_count
from target_user, target_session, target_game, disallow_jwt;

-- name: GetValidSession :one
select gs.last_pulse_at, gt.uuid as game_token_uuid
from game_session gs
    join game_token gt on gs.game_token_id = gt.id
    join game g on gs.game_id = g.id
    join users u on gs.user_id = u.id
where not exists (select * from token_disallow_list tdl where tdl.token_id = @session_token_uuid)
  and g.uuid = @game_uuid
  and u.uuid = @user_uuid
  and gs.uuid = @session_uuid
limit 1;

-- name: GetGameSessionUserProgress :many
select a.slug, ap.progress
from achievement_progress ap
join users u on ap.user_id = u.id
join achievement a on ap.achievement_id = a.id
join game g on a.game_id = g.id
where u.uuid = @user_uuid and g.uuid = @game_uuid;

-- name: UpdateGameSessionUserProgress :batchone
with target_user as (
    select id from users where users.uuid = @user_uuid
), target_achievement as (
    select a.id, a.slug, a.progress_requirement
    from achievement a
    join game g on a.game_id = g.id
    where a.slug = @achievement_slug and g.uuid = @game_uuid
)
insert into achievement_progress (user_id, achievement_id, progress)
select target_user.id, target_achievement.id, @new_progress
from target_user, target_achievement
where @new_progress <= target_achievement.progress_requirement
on conflict (user_id, achievement_id)
    do update set progress = excluded.progress
    where excluded.progress >= achievement_progress.progress
returning (select target_achievement.slug from target_achievement), achievement_progress.progress;

-- name: CreateGameSession :one
with target_game as (
    select id from game where game.uuid = @game_uuid
), target_user as (
    select id from users where users.uuid = @user_uuid
), target_game_token as (
    select id from game_token where game_token.uuid = @game_token_uuid
)
insert into game_session (game_id, user_id, game_token_id)
select target_game.id, target_user.id, target_game_token.id
from target_game, target_user, target_game_token
returning *;

-- name: HeartbeatGameSession :one
update game_session
set last_pulse_at = now()
where uuid = @session_uuid
returning last_pulse_at;
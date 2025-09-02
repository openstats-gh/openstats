-- name: FindAchievementBySlug :one
select a.*
from achievement a
     join game g on a.game_id = g.id
where a.slug = @achievement_slug
  and g.uuid = @game_uuid
limit 1;

-- name: UpsertAchievement :one
insert into achievement (game_id, slug, name, description, progress_requirement)
values ($1, $2, $3, $4, $5)
on conflict(game_id, slug)
    do update set name=excluded.name,
                  description=excluded.description,
                  progress_requirement=excluded.progress_requirement
returning case when achievement.created_at == achievement.updated_at then true else false end as upsert_was_insert;

-- name: GetUsersRarestAchievements :many
select ap.*, g.uuid game_uuid, ar.slug, ar.name, ar.description, ar.completion_percent::double precision as rarity
from achievement_progress ap
     join achievement_rarity ar on ap.achievement_id = ar.id and ap.progress >= ar.progress_requirement
     join game g on ar.game_id = g.id
     join users u on ap.user_id = u.id
where u.uuid = @user_uuid and ar.completion_percent <= @max_completion_percent::float
order by ar.completion_percent
limit $1;

-- name: GetUsersCompletedGames :many
select g.uuid as game_uuid, (select count(*) from achievement ga where ga.game_id = gc.game_id) as achievement_count, gc.*
from game_completion gc
     join users u on gc.user_id = u.id
     join game g on gc.game_id = g.id
where u.uuid = @user_uuid and gc.has_every_achievement
limit $1;

-- name: GetUserRecentAchievements :many
select d.slug as developer_slug,
       g.uuid as game_uuid,
       g.slug as game_slug,
       '' as game_name,
       a.slug as slug,
       a.name as name,
       a.description as description
from achievement_progress ap
     join achievement a on ap.achievement_id = a.id
     join users u on ap.user_id = u.id
     join game g on a.game_id = g.id
     join developer d on g.developer_id = d.id
where u.uuid = @user_uuid
  and ap.progress >= a.progress_requirement
order by ap.created_at desc
limit $1;

-- name: GetOtherUserRecentAchievements :many
select d.slug as developer_slug, g.uuid as game_uuid, g.slug as game_slug, '' as game_name, a.slug as slug, a.name as name, a.description as description, u.uuid as user_uuid, coalesce(uldn.display_name, u.slug) as user_friendly_name
from achievement_progress ap
     join achievement a on ap.achievement_id = a.id
     join users u on ap.user_id = u.id
     join game g on a.game_id = g.id
     join developer d on g.developer_id = d.id
     left outer join user_latest_display_name uldn on u.id = uldn.user_id
where u.uuid != @excluded_user_uuid
  and ap.progress >= a.progress_requirement
order by ap.created_at desc
limit $1;

-- name: GetGameAchievementsWithRarity :many
select
    ar.slug,
    ar.name,
    ar.description,
    ar.completion_percent::double precision as rarity
from achievement_rarity ar
     join game g on ar.game_id = g.id
where g.uuid = @game_uuid
order by ar.completion_percent desc;

-- name: GetRecentGameAchievements :many
select ar.slug,
       ar.name,
       ar.description,
       ar.completion_percent::double precision as rarity,
       u.uuid as user_uuid,
       u.slug as user_slug
from achievement_progress ap
     join achievement_rarity ar on ap.achievement_id = ar.id
     join game g on ar.game_id = g.id
     join users u on ap.user_id = u.id
where g.uuid = @game_uuid and ap.progress >= ar.progress_requirement
order by ap.created_at desc
limit $1;

-- name: GetRecentGameCompletions :many
select gc.unlocked_at,
       u.uuid as user_uuid,
       u.slug as user_slug
from game_completion gc
     join game g on gc.game_id = g.id
     join users u on gc.user_id = u.id
where g.uuid = @game_uuid and gc.has_every_achievement
order by gc.unlocked_at desc
limit $1;
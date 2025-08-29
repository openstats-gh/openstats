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
with achievement_rarity as (
    select a.id,
           a.slug,
           a.name,
           a.description,
           a.game_id,
           (count(*)::float / (select count(distinct gs.user_id)
                               from game_session gs
                               where gs.game_id = a.game_id)) as completion_percent
    from achievement_progress ap
         join achievement a on ap.achievement_id = a.id
    where ap.progress >= a.progress_requirement
    group by a.id
)
select ap.*, g.uuid game_uuid, ar.slug, ar.name, ar.description, ar.completion_percent::float as rarity
from achievement_progress ap
     join achievement_rarity ar on ap.achievement_id = ar.id
     join users u on ap.user_id = u.id
     join game g on ar.game_id = g.id
where u.uuid = @user_uuid and ar.completion_percent <= @max_completion_percent::float
order by ar.completion_percent
limit $1;

-- name: GetUsersCompletedGames :many
with target_user_id as (
    -- TODO: maybe this can just be a left outer join on in the inner subquery?
    select u.id from users u where u.uuid = @user_uuid
)
select *, (select count() from achievement a1 where a1.game_id = g.id) as achievement_count
from game g
where true = all (select (ap.progress >= a.progress_requirement) as completed
                  from achievement a
                       left outer join achievement_progress ap
                           on a.id = ap.achievement_id and
                              ap.user_id = target_user_id
                  where a.game_id = g.id)
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

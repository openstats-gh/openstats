-- name: FindUser :one
select * from users where users.uuid = $1 limit 1;

-- name: FindUserById :one
select * from users where users.id = @user_id limit 1;

-- name: FindUserBySlug :one
select * from users where slug = $1 limit 1;

-- name: GetUserUuid :one
select uuid from users where slug = $1 limit 1;

-- name: GetUserSessionProfile :one
select
    u.uuid,
    u.slug,
    coalesce(uldn.display_name, ''),
    u.created_at,
    ua.uuid as avatar_uuid,
    ua.blurhash as avatar_blurhash
from users u
     left outer join user_latest_display_name uldn on u.id = uldn.user_id
     left outer join user_avatar ua on u.id = ua.user_id
where u.uuid = @user_uuid
limit 1;

-- name: GetUserWithName :one
select u.uuid, u.created_at, u.slug, coalesce(uldn.display_name, '')
from users u
    left outer join user_latest_display_name uldn on u.id = uldn.user_id
where u.uuid = @user_uuid
limit 1;

-- name: UpdateSessionProfile :exec
with target_user as (
    select u1.id, u1.slug as old_slug, uldn.display_name as latest_display_name
    from users u1
         left outer join user_latest_display_name uldn on u1.id = uldn.user_id
    where u1.uuid = @uuid
), _ as (
    update users u2
    set slug = cast(sqlc.narg(new_slug) as text)
    from target_user
    where u2.id = target_user.id and cast(sqlc.narg(new_slug) as text) is not null and cast(sqlc.narg(new_slug) as text) != target_user.old_slug
)
insert into user_display_name (user_id, display_name)
select target_user.id, cast(sqlc.narg(new_display_name) as text)
from target_user
where cast(sqlc.narg(new_display_name) as text) is not null and cast(sqlc.narg(new_display_name) as text) != target_user.latest_display_name;

-- name: FindUserBySlugWithPassword :one
select u.*, up.encoded_hash
from users u
     join user_password up on u.id = up.user_id
where u.slug = $1
limit 1;

-- name: AddUser :one
insert into users (slug) values ($1) returning *;

-- name: AddUserPassword :exec
insert into user_password(user_id, encoded_hash) values ($1, $2);

-- name: AddUserSlugHistory :exec
insert into user_slug_history(user_id, slug) values ($1, $2);

-- name: AddUserEmail :exec
insert into user_email(user_id, email) values ($1, $2);

-- name: AddUserDisplayName :exec
insert into user_display_name(user_id, display_name) values ($1, $2);

-- name: AllUsersWithDisplayNames :many
select u.*, uldn.display_name
from users u
    left outer join user_latest_display_name uldn on u.id = uldn.user_id;

-- name: GetUserDisplayNames :many
select *
from user_display_name
where user_id = $1;

-- name: GetUserLatestDisplayName :one
select *
from user_display_name udn
where udn.user_id = $1
order by udn.created_at desc
limit 1;

-- name: GetUserEmails :many
select *
from user_email
where user_id = $1;

-- name: GetUserDevelopers :many
select d.slug, d.created_at, dm.created_at as joined_at
from developer_member dm
     join developer d on dm.developer_id = d.id
where dm.user_id = $1;

-- name: GetUserRecentAchievements :many
select d.slug as developer_slug, g.slug as game_slug, '' as game_name, a.slug as slug, a.name as name, a.description as description
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
select d.slug as developer_slug, g.slug as game_slug, '' as game_name, a.slug as slug, a.name as name, a.description as description, u.uuid as user_uuid, coalesce(uldn.display_name, u.slug) as user_friendly_name
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

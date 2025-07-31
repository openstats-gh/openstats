-- name: FindUser :one
select * from users where id = $1 limit 1;

-- name: FindUserBySlug :one
select *
from users
where slug = $1
limit 1;

-- name: FindUserByLookupId :one
select *
from users
where lookup_id = $1
limit 1;

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
where u.slug = @user_slug
  and ap.progress >= a.progress_requirement
order by ap.created_at desc
limit $1;

-- name: GetOtherUserRecentAchievements :many
select d.slug as developer_slug, g.slug as game_slug, '' as game_name, a.slug as slug, a.name as name, a.description as description, u.slug as user_slug, uldn.display_name as user_display_name
from achievement_progress ap
     join achievement a on ap.achievement_id = a.id
     join users u on ap.user_id = u.id
     join game g on a.game_id = g.id
     join developer d on g.developer_id = d.id
     left outer join user_latest_display_name uldn on u.id = uldn.user_id
where u.slug != @excluded_user_slug
  and ap.progress >= a.progress_requirement
order by ap.created_at desc
limit $1;

-- Batch Inserts:

-- name: FindUserUUIDsBySlugs :many
select lookup_id from users where slug = any(sqlc.slice(slugs));

-- name: AddUsers :copyfrom
insert into users (slug) values ($1);

-- name: AddUserSlugHistories :copyfrom
insert into user_slug_history (user_id, slug) values ($1, $2);

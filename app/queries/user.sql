-- name: FindUser :one
select *
from user
where id = ?
limit 1;

-- name: FindUserBySlug :one
select *
from user
where slug = ?
limit 1;

-- name: FindUserBySlugWithPassword :one
select user.*, user_password.encoded_hash
from user
     join user_password on user.id = user_password.user_id
where user.slug = ?
  and user.deleted_at is null
limit 1;

-- name: CreateUser :one
insert into user (updated_at, slug)
values (datetime('now'), ?)
returning *;

-- name: CreateUserPassword :exec
insert into user_password (updated_at, user_id, encoded_hash)
values (datetime('now'), ?, ?);

-- name: AddUserSlugRecord :exec
insert into user_slug (user_id, slug)
values (?, ?);

-- name: AddUserEmail :exec
insert into user_email (updated_at, user_id, email)
values (datetime('now'), ?, ?);

-- name: AddUserDisplayName :exec
insert into user_display_name (user_id, display_name)
values (?, ?);

-- name: AllUsersWithDisplayNames :many
select u.*, udn1.display_name
from user u
     left outer join user_display_name udn1 on u.id = udn1.user_id and udn1.deleted_at is null
     left outer join user_display_name udn2 on u.id = udn2.user_id and udn2.deleted_at is null and
                                               (udn1.created_at < udn2.created_at or
                                                (udn1.created_at = udn2.created_at and udn1.id < udn2.id))
where udn2.id is null;

-- name: GetUserDisplayNames :many
select *
from user_display_name
where user_id = ?;

-- name: GetUserLatestDisplayName :one
select *
from user_display_name udn
where udn.user_id = ? and udn.deleted_at is null
order by udn.created_at desc
limit 1;

-- name: GetUserEmails :many
select *
from user_email
where user_id = ?;

-- name: GetUserDevelopers :many
select d.slug, d.created_at, d.deleted_at, dm.created_at as joined_at, dm.deleted_at as left_at
from developer_member dm
     join developer d on dm.developer_id = d.id
where dm.user_id = ?;

-- name: GetUserRecentAchievements :many
select d.slug as developer_slug, g.slug as game_slug, '' as game_name, a.slug as slug, a.name as name, a.description as description
from achievement_progress ap
     join achievement a on ap.achievement_id = a.id
     join user u on ap.user_id = u.id
     join game g on a.game_id = g.id
     join developer d on g.developer_id = d.id
where u.slug = sqlc.arg(user_slug)
  and ap.progress >= a.progress_requirement
  and u.deleted_at is null
  and g.deleted_at is null
  and a.deleted_at is null
  and ap.deleted_at is null
order by ap.created_at desc
limit ?;

-- name: GetOtherUserRecentAchievements :many
select d.slug as developer_slug, g.slug as game_slug, '' as game_name, a.slug as slug, a.name as name, a.description as description, u.slug as user_slug, udn1.display_name as user_display_name
from achievement_progress ap
     join achievement a on ap.achievement_id = a.id
     join user u on ap.user_id = u.id
     join game g on a.game_id = g.id
     join developer d on g.developer_id = d.id
     left outer join user_display_name udn1 on u.id = udn1.user_id and udn1.deleted_at is null
     left outer join user_display_name udn2 on u.id = udn2.user_id and udn2.deleted_at is null and
                                               (udn1.created_at < udn2.created_at or
                                                (udn1.created_at = udn2.created_at and udn1.id < udn2.id))
where u.slug != sqlc.arg(excluded_user_slug)
  and ap.progress >= a.progress_requirement
  and u.deleted_at is null
  and g.deleted_at is null
  and a.deleted_at is null
  and ap.deleted_at is null
order by ap.created_at desc
limit ?;
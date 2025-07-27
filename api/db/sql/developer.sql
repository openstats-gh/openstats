-- name: AllDevelopers :many
select * from developer;

-- name: FindDeveloperBySlug :one
select * from developer where slug = $1 limit 1;

-- name: GetDeveloperMembers :many
select u.slug, uldn.display_name, dm.created_at as joined_at
from users u
join developer_member dm on u.id = dm.user_id
left outer join user_latest_display_name uldn on u.id = uldn.user_id
where dm.developer_id = $1;

-- name: GetDeveloperGames :many
select game.slug, game.created_at from game where developer_id = $1;

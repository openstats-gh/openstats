-- name: AllDevelopers :many
select * from developer;

-- name: FindDeveloperBySlug :one
select * from developer where slug = $1 limit 1;

-- name: GetDeveloperMembers :many
select u.slug, udn1.display_name, dm.created_at as joined_at
from users u
join developer_member dm on u.id = dm.user_id
left outer join user_display_name udn1 on u.id = udn1.user_id
left outer join user_display_name udn2 on u.id = udn2.user_id and
                                          (udn1.created_at < udn2.created_at or
                                           (udn1.created_at = udn2.created_at and udn1.id < udn2.id))
where dm.developer_id = $1;

-- name: GetDeveloperGames :many
select game.slug, game.created_at from game where developer_id = $1;

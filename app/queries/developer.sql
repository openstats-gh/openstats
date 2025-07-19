-- name: AllDevelopers :many
select *
from developer;

-- name: FindDeveloperBySlug :one
select *
from developer
where slug = ?
limit 1;

-- name: GetDeveloperMembers :many
select user.slug
from user
     join developer_member on user.id = developer_member.user_id
where developer_member.developer_id = ?;

-- name: GetDeveloperGames :many
select game.slug
from game
where developer_id = ?;

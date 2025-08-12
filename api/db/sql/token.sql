-- name: CreateToken :one
insert into token (issuer, subject, audience, expires_at, not_before, issued_at)
values ($1, $2, $3, $4, $5, $6)
returning *;

-- name: DisallowToken :exec
insert into token_disallow_list (token_id) values ($1);

-- name: FindUserGameTokens :many
select gt.created_at, gt.expires_at, gt.uuid, gt.comment, g.uuid as game_uuid, g.slug as game_slug, d.slug as developer_slug
from game_token gt
join game g on gt.game_id = g.id
join developer d on g.developer_id = d.id -- TODO: developer_latest_display_name
join users u on gt.user_id = u.id
where u.uuid = @user_uuid and gt.expires_at > now();

-- name: CreateGameToken :one
with target_user as (
    select id from users where users.uuid = @user_uuid
), target_game as (
    select g.id, g.uuid, g.slug, d.slug as developer_slug
    from game g
    join developer d on g.developer_id = d.id
    where g.uuid = @game_uuid
)
insert into game_token (expires_at, comment, game_id, user_id)
values (@expires_at, @comment, (select id from target_game), (select id from target_user))
returning
    game_token.uuid,
    game_token.expires_at,
    game_token.created_at,
    game_token.comment,
    (select uuid as game_uuid from target_game),
    (select slug as game_slug from target_game),
    (select developer_slug from target_game);

-- name: ExpireToken :execrows
delete
from game_token gt
where gt.user_id = (select u.id from users u where u.uuid = @user_uuid)
  and gt.uuid = @uuid;

-- name: FindTokenWithUser :one
select u.uuid as user_uuid, g.uuid as game_uuid
from game_token gt
join game g on gt.game_id = g.id
join users u on gt.user_id = u.id
where gt.uuid = @uuid and gt.expires_at > now()
limit 1;
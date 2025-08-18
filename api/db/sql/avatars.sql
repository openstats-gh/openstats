-- name: AddUserAvatar :one
insert into user_avatar (user_id, blurhash)
select u.id, @blurhash from users u where u.uuid = @user_uuid
returning *;

-- name: AddDeveloperAvatar :one
insert into developer_avatar (developer_id, blurhash)
select d.id, @blurhash from developer d where d.uuid = @developer_uuid
returning *;

-- name: AddGameAvatar :one
insert into game_avatar (game_id, blurhash)
select g.id, @blurhash from game g where g.uuid = @game_uuid
returning *;

-- name: AddAchievementAvatar :one
insert into achievement_avatar (achievement_id, blurhash)
select a.id, @blurhash
from game g
join achievement a on g.id = a.game_id
where g.uuid = @game_uuid and a.slug = @achievement_slug
returning *;

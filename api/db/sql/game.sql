-- name: AllGames :many
select game.*, developer.slug as developer_slug
from game
     join developer on game.developer_id = developer.id;

-- name: FindGameBySlug :one
select game.*
from game
     join developer on game.developer_id = developer.id
where game.slug = @game_slug and developer.slug = @dev_slug;

-- name: GetGameAchievements :many
select * from achievement where game_id = $1;

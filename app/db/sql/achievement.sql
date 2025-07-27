-- name: FindAchievementBySlug :one
select a.*
from achievement a
     join game g on a.game_id = g.id
     join developer d on g.developer_id = d.id
where a.slug = @ach_slug
  and d.slug = @dev_slug
  and g.slug = @game_slug
limit 1;

-- name: UpsertAchievement :one
insert into achievement (game_id, slug, name, description, progress_requirement)
values ($1, $2, $3, $4, $5)
on conflict(game_id, slug)
    do update set name=excluded.name,
                  description=excluded.description,
                  progress_requirement=excluded.progress_requirement
returning case when achievement.created_at == achievement.updated_at then true else false end as upsert_was_insert;
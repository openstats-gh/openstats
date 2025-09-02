create or replace view achievement_rarity as
select a.*,
       count(*)::float as completion_count,
       (count(*)::float / (select count(distinct gs.user_id)
                           from game_session gs
                           where gs.game_id = a.game_id))::float as completion_percent
from achievement_progress ap
     join achievement a on ap.achievement_id = a.id
where ap.progress >= a.progress_requirement
group by a.id;

create or replace view game_completion as
select g.id as game_id,
       u.id as user_id,
       (select ap1.created_at
        from achievement_progress ap1
             join achievement a1 on ap1.achievement_id = a1.id
        where a1.game_id = g.id and ap1.user_id = u.id
        order by ap1.created_at
        limit 1) as unlocked_at,
       count(*) as unlock_count,
       count(*) = (select count(*) from achievement ga where ga.game_id = g.id) as has_every_achievement
from achievement a
     join game g on a.game_id = g.id
     join achievement_progress ap on a.id = ap.achievement_id and ap.progress >= a.progress_requirement
     join users u on ap.user_id = u.id
group by g.id, u.id;

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
       count(*) as unlock_count,
       count(*) = (select count(*) from achievement ga where ga.game_id = g.id) as has_every_achievement
from achievement a
     join game g on a.game_id = g.id
     join achievement_progress ap on a.id = ap.achievement_id and ap.progress >= a.progress_requirement
     join users u on ap.user_id = u.id
group by g.id, u.id;

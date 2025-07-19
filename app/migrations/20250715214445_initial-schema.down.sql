drop table if exists user;
drop index if exists user_unique_slug;
drop table if exists user_slug;
drop table if exists user_email;
drop table if exists user_display_name;
drop table if exists user_password;
drop table if exists developer;
drop index if exists developer_unique_slug;
drop table if exists developer_member;
drop table if exists developer_slug;
drop table if exists game;
drop index if exists game_unique_slug_per_dev;
drop table if exists game_slug;
drop table if exists achievement;
drop index if exists achievement_unique_slug_per_game;
drop table if exists achievement_progress;

-- TODO: any other indices/constraints?
create table if not exists user
(
    id         integer primary key,
    created_at datetime not null default (datetime('now')),
    updated_at datetime not null,
    deleted_at datetime,
    slug       text     not null
);

create unique index user_unique_slug on user(slug);

create table if not exists user_slug
(
    id         integer primary key,
    created_at datetime not null default (datetime('now')),
    deleted_at datetime,
    user_id    integer not null references user (id),
    slug       text     not null
);

create table if not exists user_email
(
    id           integer primary key,
    created_at   datetime not null default (datetime('now')),
    updated_at   datetime not null,
    deleted_at   datetime,
    user_id      integer not null references user (id),
    email        text     not null,
    confirmed_at datetime
);

create table if not exists user_display_name
(
    id           integer primary key,
    created_at   datetime not null default (datetime('now')),
    deleted_at   datetime,
    user_id      integer not null references user (id),
    display_name text     not null
);

create table if not exists user_password
(
    id           integer primary key,
    created_at   datetime not null default (datetime('now')),
    updated_at   datetime not null,
    user_id      integer not null references user (id),
    encoded_hash text     not null
);

create table if not exists developer
(
    id         integer primary key,
    created_at datetime not null default (datetime('now')),
    updated_at datetime not null,
    deleted_at datetime,
    slug       text     not null
);

create unique index developer_unique_slug on developer(slug);

create table if not exists developer_member
(
    id           integer primary key,
    created_at   datetime not null default (datetime('now')),
    deleted_at   datetime,
    user_id      integer not null references user (id),
    developer_id integer not null references developer (id)
);

create table if not exists developer_slug
(
    id           integer primary key,
    created_at   datetime not null default (datetime('now')),
    deleted_at   datetime,
    developer_id integer not null references developer (id),
    slug         text     not null
);

create table if not exists game
(
    id           integer primary key,
    created_at   datetime not null default (datetime('now')),
    updated_at   datetime not null,
    deleted_at   datetime,
    developer_id integer not null references developer (id),
    slug         text     not null
);

create unique index game_unique_slug_per_dev on game(developer_id, slug);

create table if not exists game_slug
(
    id         integer primary key,
    created_at datetime not null default (datetime('now')),
    deleted_at datetime,
    game_id    integer not null references game (id),
    slug       text     not null
);

create table if not exists achievement
(
    id                   integer primary key,
    created_at           datetime not null default (datetime('now')),
    updated_at           datetime not null,
    deleted_at           datetime,
    game_id              integer not null references game (id),
    slug                 text     not null,
    name                 text     not null,
    description          text     not null,
    progress_requirement integer  not null
);

create unique index achievement_unique_slug_per_game on achievement(game_id, slug);

create table if not exists achievement_progress
(
    created_at     datetime not null default (datetime('now')),
    updated_at     datetime not null,
    deleted_at     datetime,
    user_id        integer not null references user (id),
    achievement_id integer not null references achievement (id),
    progress       integer  not null,

    primary key (user_id, achievement_id)
);

-- TODO: any other indices/constraints?
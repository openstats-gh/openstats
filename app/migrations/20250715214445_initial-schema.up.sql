create extension if not exists moddatetime;
create extension if not exists pgcrypto;

/*
gen_ulid generates a ULID-based UUID: https://github.com/ulid/spec
based on implementation here: https://web.archive.org/web/20250525070451/https://blog.daveallie.com/ulid-primary-keys/

N.B.
    this implementation of gen_ulid is very slow compared to gen_random_uuid()

TODO: i'd really like to replace this with pg-ulid: https://github.com/andrielfn/pg-ulid
*/
create or replace function gen_ulid() returns uuid
as
$$
select (lpad(to_hex(floor(extract(epoch from clock_timestamp()) * 1000)::bigint), 12, '0') ||
        encode(gen_random_bytes(10), 'hex'))::uuid;
$$ language sql;

/*
Example soft-delete query using deleted_record table:

with deleted AS (
    delete from users
    where id = ?
    returning *
)
insert into deleted_record(source_table, source_id, data)
select 'users', id, to_jsonb(deleted.*)
from deleted
returning *;
*/
create table if not exists deleted_record
(
    id           uuid primary key     default gen_ulid(),
    deleted_at   timestamptz not null default now(),
    source_table text        not null,
    source_id    text        not null,
    data         jsonb       not null
);

create table if not exists users
(

    id         serial primary key,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    slug       text        not null unique
);
create or replace trigger users_moddatetime
    before update
    on users
    for each row
execute function moddatetime(updated_at);

comment on table users is 'openstats users. table is plural to avoid name collision with pg `user` keyword.';

create table if not exists user_slug_history
(
    id         serial primary key,
    created_at timestamptz not null default now(),
    user_id    integer     not null references users,
    slug       text        not null
);

create table if not exists user_email
(
    id           serial primary key,
    created_at   timestamptz not null default now(),
    updated_at   timestamptz not null default now(),
    user_id      integer     not null references users,
    email        text        not null,
    confirmed_at timestamptz
);
create or replace trigger user_email_moddatetime
    before update
    on user_email
    for each row
execute function moddatetime(updated_at);

create table if not exists user_display_name
(
    id           serial primary key,
    created_at   timestamptz not null default now(),
    user_id      integer     not null references users,
    display_name text        not null
);

create table if not exists user_password
(
    id           serial primary key,
    created_at   timestamptz not null default now(),
    updated_at   timestamptz not null default now(),
    user_id      integer     not null references users,
    encoded_hash text        not null
);
create or replace trigger user_password_moddatetime
    before update
    on user_password
    for each row
execute function moddatetime(updated_at);

create table if not exists developer
(
    id         serial primary key,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    slug       text        not null unique
);
create or replace trigger developer_moddatetime
    before update
    on developer
    for each row
execute function moddatetime(updated_at);

create table if not exists developer_member
(
    id           serial primary key,
    created_at   timestamptz not null default now(),
    user_id      integer     not null references users,
    developer_id integer     not null references developer
);

create table if not exists developer_slug_history
(
    id           serial primary key,
    created_at   timestamptz not null default now(),
    developer_id integer     not null references developer,
    slug         text        not null
);

create table if not exists developer_display_name
(
    id           serial primary key,
    created_at   timestamptz not null default now(),
    developer_id integer     not null references developer,
    display_name text        not null
);

create table if not exists game
(
    id           serial primary key,
    created_at   timestamptz not null default now(),
    updated_at   timestamptz not null default now(),
    developer_id integer     not null references developer,
    slug         text        not null,

    unique (developer_id, slug)
);
create or replace trigger game_moddatetime
    before update
    on game
    for each row
execute function moddatetime(updated_at);

create table if not exists achievement
(
    id                   serial primary key,
    created_at           timestamptz not null default now(),
    updated_at           timestamptz not null default now(),
    game_id              integer     not null references game,
    slug                 text        not null,
    name                 text        not null,
    description          text        not null,
    progress_requirement integer     not null
);

create table if not exists achievement_progress
(
    created_at     timestamptz not null default now(),
    updated_at     timestamptz not null default now(),
    user_id        integer     not null references users,
    achievement_id integer     not null references achievement,
    progress       integer     not null,

    primary key (user_id, achievement_id)
);

/*
game tokens are JWTs generated by users which include these claims:
    iss: a resource path to the object which owns/created the slug
         e.g user/some-user-slug
             developer/some-developer-slug
             developer/some-developer-slug/game/some-game-slug
    sub: a resource path to the authorized user
         e.g. user/some-user-slug
    aud: a resource path to the game which this token is intended for
         e.g. developer/some-developer-slug/game/some-game-slug
    exp: an expiration timestamp, this is chosen by the user when they generate the token
    nbf: always the timestamp that the token was created at
    iat: always the timestamp that the token was created at
    jti: a ULID 

the claims are used to verify that the submitter has permission to submit achievement progress and 
game stats for a particular user.

the token table itself just stores information about issued tokens, it is not used for claims validation,
authentication, or authorization. Only the private key & JWT claims are used for those.
*/
create table if not exists token
(
    id         uuid primary key default gen_ulid(),
    issuer     text        not null,
    subject    text        not null,
    audience   text        not null,
    expires_at timestamptz not null,
    not_before timestamptz not null,
    issued_at  timestamptz not null
);

/*
the token_disallow_list table lists JWTs which have been manually expired/disallowed. Entries
in this database are eventually deleted after now() has surpassed expire_at sufficiently.
*/
create table if not exists token_disallow_list
(
    token_id   uuid primary key references token,
    created_at timestamptz not null default now()
);

create table if not exists game_session
(
    id            uuid primary key     default gen_ulid(),
    created_at    timestamptz not null default now(),
    token_id      uuid references token,
    -- last time a pulse was received by the game
    last_pulse_at timestamptz not null default now(),
    game_id       integer     not null references game,
    user_id       integer     not null references users
);

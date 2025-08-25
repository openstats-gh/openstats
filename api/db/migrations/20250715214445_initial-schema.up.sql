create extension if not exists moddatetime;
create extension if not exists pgcrypto;

/*
gen_uuid_v7 generates a Version 7 UUID
based on implementation here: https://postgresql.verite.pro/blog/2024/07/15/uuid-v7-pure-sql.html

TODO: replace with pgxn uuidv7 extension or upgrade to postgres 18

 */
create or replace function gen_uuid_v7() returns uuid
as
$$
    -- Replace the first 48 bits of a uuidv4 with the current
    -- number of milliseconds since 1970-01-01 UTC
    -- and set the "ver" field to 7 by setting additional bits
    select encode(
       set_bit(
           set_bit(
               overlay(uuid_send(gen_random_uuid()) placing
                   substring(int8send((extract(epoch from clock_timestamp())*1000)::bigint) from 3)
                   from 1 for 6),
               52, 1),
           53, 1), 'hex')::uuid;
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

alternatively, we can create a function:

    create function deleted_record_insert() returns trigger
        language plpgsql
    as $$
        begin
            execute 'insert into deleted_record (data, object_id, table_name) values ($1, $2, $3)'
            using to_jsonb(old.*), old.id, TG_TABLE_NAME;

            return old;
        end;
    $$;

and add an after delete trigger:

    create trigger deleted_record_insert after delete on my_table
        for each row execute function deleted_record_insert();

*/
create table if not exists deleted_record
(
    id           uuid primary key     default gen_uuid_v7(),
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
    uuid       uuid        not null default gen_uuid_v7() unique,
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

create or replace view user_latest_email as
select coalesce(ue2.id, ue1.id) as id,
       coalesce(ue2.created_at, ue1.created_at) as created_at,
       coalesce(ue2.updated_at, ue1.updated_at) as updated_at,
       coalesce(ue2.user_id, ue1.user_id) as user_id,
       coalesce(ue2.email, ue1.email) as email,
       coalesce(ue2.confirmed_at, ue1.confirmed_at) as confirmed_at
from users u
    left outer join user_email ue1 on u.id = ue1.user_id
    left outer join user_email ue2 on u.id = ue2.user_id and
                                      (ue1.created_at < ue2.created_at or
                                       (ue1.created_at = ue2.created_at and ue1.id < ue2.id));

create table if not exists user_display_name
(
    id           serial primary key,
    created_at   timestamptz not null default now(),
    user_id      integer     not null references users,
    display_name text        not null
);

create index if not exists user_display_name_created_at on user_display_name(created_at);

create or replace view user_latest_display_name as
select coalesce(udn2.id, udn1.id) as id,
       coalesce(udn2.created_at, udn1.created_at) as created_at,
       coalesce(udn2.user_id, udn1.user_id) as user_id,
       coalesce(udn2.display_name, udn1.display_name) as display_name
from users u
     left outer join user_display_name udn1 on u.id = udn1.user_id
     left outer join user_display_name udn2 on u.id = udn2.user_id and
                                               (udn1.created_at < udn2.created_at or
                                                (udn1.created_at = udn2.created_at and udn1.id < udn2.id));

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

create table if not exists user_avatar
(
    id         serial primary key,
    created_at timestamptz not null default now(),
    uuid       uuid        not null unique default gen_uuid_v7(),
    blurhash   text not null,
    user_id    integer not null references users
);

create table if not exists developer
(
    id         serial primary key,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    uuid       uuid        not null unique default gen_uuid_v7(),
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

create table if not exists developer_avatar
(
    id           serial primary key,
    created_at   timestamptz not null default now(),
    uuid         uuid        not null unique default gen_uuid_v7(),
    blurhash     text not null,
    developer_id integer not null references developer
);

create table if not exists game
(
    id           serial primary key,
    created_at   timestamptz not null default now(),
    updated_at   timestamptz not null default now(),
    developer_id integer     not null references developer,
    uuid         uuid        not null unique default gen_uuid_v7(),
    slug         text        not null,

    unique (developer_id, slug)
);
create or replace trigger game_moddatetime
    before update
    on game
    for each row
execute function moddatetime(updated_at);

create table if not exists game_avatar
(
    id         serial primary key,
    created_at timestamptz not null default now(),
    uuid       uuid        not null unique default gen_uuid_v7(),
    blurhash   text not null,
    game_id    integer not null references game
);

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

create table if not exists achievement_avatar
(
    id             serial primary key,
    created_at     timestamptz not null default now(),
    uuid           uuid        not null unique default gen_uuid_v7(),
    blurhash       text not null,
    achievement_id integer not null references achievement
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
    id         uuid primary key default gen_uuid_v7(),
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

create table if not exists game_token
(
    id serial primary key,
    created_at timestamptz not null default now(),
    expires_at timestamptz not null,
    uuid     uuid not null unique default gen_uuid_v7(),
    comment  text not null,
    user_id  int not null references users,
    game_id  int not null references game
);

create table if not exists game_session
(
    id            uuid primary key     default gen_uuid_v7(),
    created_at    timestamptz not null default now(),
    uuid          uuid        not null unique default gen_uuid_v7(),
    game_id       integer     not null references game,
    user_id       integer     not null references users,
    game_token_id integer     not null references game_token,
    -- last time a pulse was received by the game
    last_pulse_at timestamptz not null default now()
);


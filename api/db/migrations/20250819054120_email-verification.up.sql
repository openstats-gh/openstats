create table if not exists user_email_confirmation(
    id            serial      primary key,
    created_at    timestamptz not null default now(),
    user_email_id integer     not null references user_email,
    code          text        not null,
    expires_at    timestamptz not null,
    confirmed_at  timestamptz null
);

create index if not exists user_email_confirmation__expires_at__code
    on user_email_confirmation(expires_at, code);

-- converting any existing data in user_email to the new user_email_confirmation table.
-- in order to avoid an exploit where someone could pass an empty code and verify their email,
-- we always set the expiration to 1 minute ago so any old entries are expired after this migration.
insert into user_email_confirmation(created_at, user_email_id, code, expires_at, confirmed_at)
select ue.created_at, ue.id, '', now() - '1 minute'::interval, ue.confirmed_at
from user_email ue;

-- remove updated_at and confirmed_at from user_latest_email view
drop view if exists user_latest_email;
create view user_latest_email as
select coalesce(ue2.id, ue1.id) as id,
       coalesce(ue2.created_at, ue1.created_at) as created_at,
       coalesce(ue2.user_id, ue1.user_id) as user_id,
       coalesce(ue2.email, ue1.email) as email
from users u
     left outer join user_email ue1 on u.id = ue1.user_id
     left outer join user_email ue2 on u.id = ue2.user_id and
                                       (ue1.created_at < ue2.created_at or
                                        (ue1.created_at = ue2.created_at and ue1.id < ue2.id));

alter table user_email drop column confirmed_at;
alter table user_email drop column updated_at;

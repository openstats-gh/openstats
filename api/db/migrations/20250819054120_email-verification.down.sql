alter table user_email add column if not exists updated_at timestamptz not null default now();
alter table user_email add column if not exists confirmed_at timestamptz null;

-- add back updated_at and confirmed_at to user_email based on some assumptions about user_email_confirmation
update user_email as ue
set updated_at = uec.updated_at,
    confirmed_at = uec.confirmed_at
from (select user_email_id,
             coalesce(confirmed_at, created_at) as updated_at,
             confirmed_at
      from user_email_confirmation
) as uec
where ue.id = uec.user_email_id;

-- add back updated_at and confirmed_at into user_latest_email view
drop view if exists user_latest_email;
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

drop index if exists user_email_confirmation__expires_at__code;
drop table if exists user_email_confirmation;
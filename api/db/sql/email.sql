-- name: AddOrGetUserEmail :one
insert into user_email (user_id, email, otp_secret)
values (@user_id, @email, @otp_secret)
on conflict (user_email_unique_idx) do nothing
returning *;

-- name: AddOrGetUserEmailByUuid :one
with target_user as (
    select u.id from users u where u.uuid = @user_uuid
)
insert into user_email (user_id, email, otp_secret)
select tu.id, @email, @otp_secret from target_user tu
on conflict (user_email_unique_idx) do nothing
returning *;

-- name: GetUserEmail :one
select *
from user_email
where user_id = @user_id and email = @email;

-- name: GetUserEmails :many
select *
from user_email
where user_id = @user_id;

-- name: ConfirmEmail :one
update user_email
set confirmed_at = now()
where user_id = @user_id and email = @email
returning *;

-- name: RemoveEmail :one
delete
from user_email
where user_id = @user_id and email = @email
returning *;

-- name: GetSlugsByEmail :many
select u.slug
from users u
     join user_email ue on u.id = ue.user_id
where ue.email = @email and ue.confirmed_at is not null;
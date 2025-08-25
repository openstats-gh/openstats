-- name: AddOrGetUserEmail :one
with target_user as (
    select u.id from users u where u.uuid = @user_uuid
)
insert into user_email (user_id, email, otp_secret)
select tu.id, @email, @otp_secret from target_user tu
on conflict (user_email_unique_idx) do nothing
returning *;

-- name: GetUserEmail :one
select ue.*
from user_email ue
     join users u on ue.user_id = u.id
where u.uuid = @user_uuid and ue.email = @email;

-- name: GetUserEmails :many
select ue.*
from user_email ue
     join users u on ue.user_id = u.id
where u.uuid = @user_uuid;
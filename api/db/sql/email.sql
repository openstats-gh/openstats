<<<<<<< HEAD
-- name: AddOrGetUserEmail :one
insert into user_email(user_id, email)
values (@user_id, @email)
on conflict (user_id) do nothing
returning *;

-- name: AddUserEmailByUuid :exec
with target_user as (
    select u.id from users u where u.uuid = @user_uuid
)
insert into user_email(user_id, email)
select tu.id, @email from target_user tu;

-- name: GetUserEmail :one
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

-- name: FindUserEmailBySlug :one
select ue.user_id, ue.email
from user_email ue
     join users u on ue.user_id = u.id
where u.slug = @slug and ue.confirmed_at is not null;
=======
-- name: AddOrGetUserEmail :one
insert into user_email(user_id, email)
values (@user_id, @email)
on conflict (user_id) do nothing
returning *;

-- name: AddUserEmailByUuid :exec
with target_user as (
    select u.id from users u where u.uuid = @user_uuid
)
insert into user_email(user_id, email)
select tu.id, @email from target_user tu;

-- name: GetUserEmail :one
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

-- name: FindUserEmailBySlug :one
select ue.user_id, ue.email
from user_email ue
     join users u on ue.user_id = u.id
where u.slug = @slug and ue.confirmed_at is not null;
>>>>>>> main

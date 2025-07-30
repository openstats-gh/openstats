-- name: CreateToken :one
insert into token (issuer, subject, audience, expires_at, not_before, issued_at)
values ($1, $2, $3, $4, $5, $6)
returning *;

-- name: DisallowToken :exec
insert into token_disallow_list (token_id) values ($1);

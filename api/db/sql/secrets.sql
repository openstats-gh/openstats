-- name: SecretRead :one
select value from secret where path = @path and key = @key;

-- name: SecretCreate :exec
insert into secret(path, key, value)
values (@path, @key, @value);

-- name: SecretWrite :exec
update secret
    set value = @value
where path = @path and key = @key;


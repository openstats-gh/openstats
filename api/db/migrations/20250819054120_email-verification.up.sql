-- alter table user_email add column otp_secret text not null default digest(gen_random_bytes(128), 'sha512');
create unique index if not exists user_email_unique_idx on user_email(user_id);

-- we use a dedicated secrets table to make it more clear when we're pulling a secret from the database
create table if not exists secret(
    id    serial primary key,
    path  text not null,
    key   text not null,
    value text not null,

    unique(path, key)
);

-- all users should have an HMAC secret by default
-- this secret is used for TOTP 2FA when:
--    - resetting password
--    - verifying emails
--
-- the private.* secrets are private, and will never be shared with anyone - including the user. If external TOTP MFA is
-- supported in the future, then we should provide a dedicated 'shared.user.mfa-hmac' secret for each user, which is
-- shared exclusively with the user.
insert into secret(path, key, value)
select 'private.user.2fa-hmac', id::text, encode(gen_random_bytes(64), 'hex') from users;

alter table user_email add column otp_secret text not null default digest(gen_random_bytes(128), 'sha512');
create unique index if not exists user_email_unique_idx on user_email(user_id, email);

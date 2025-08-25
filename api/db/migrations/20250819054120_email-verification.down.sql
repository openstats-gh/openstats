drop index if exists user_email_unique_idx;
alter table user_email drop column if exists otp_secret;

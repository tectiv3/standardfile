PRAGMA foreign_keys=OFF;
BEGIN TRANSACTION;
CREATE TABLE IF NOT EXISTS "items" (
    "uuid" varchar(36) primary key NULL,
    "user_uuid" varchar(36) NOT NULL,
    "content" blob NOT NULL,
    "content_type" varchar(255) NOT NULL,
    "enc_item_key" varchar(255) NOT NULL,
    "auth_hash" varchar(255) NOT NULL,
    "deleted" integer(1) NOT NULL DEFAULT 0,
    "created_at" timestamp(2) NOT NULL,
    "updated_at" timestamp(2) NOT NULL);
CREATE TABLE IF NOT EXISTS "users" ("uuid" varchar(36) primary key NULL, "email" varchar(255) NOT NULL, "password" varchar(255) NOT NULL, "pw_func" varchar(255) NOT NULL DEFAULT "pbkdf2", "pw_alg" varchar(255) NOT NULL DEFAULT "sha512", "pw_cost" integer NOT NULL DEFAULT 5000, "pw_key_size" integer NOT NULL DEFAULT 512, "pw_nonce" varchar(255) NOT NULL, "created_at" timestamp NOT NULL, "updated_at" timestamp NOT NULL);
CREATE INDEX IF NOT EXISTS user_uuid ON items (user_uuid);
CREATE INDEX IF NOT EXISTS user_content on items (user_uuid, content_type);
CREATE INDEX IF NOT EXISTS updated_at on items (updated_at);
CREATE INDEX IF NOT EXISTS email on users (email);
COMMIT;

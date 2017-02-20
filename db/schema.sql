PRAGMA foreign_keys=OFF;
BEGIN TRANSACTION;
CREATE TABLE "items" ("uuid" varchar(36) primary key NULL, "content" blob NULL, "content_type" varchar(255) NULL, "enc_item_key" varchar(255) NULL, "auth_hash" varchar(255) NULL, "user_uuid" varchar(36) NULL, "created_at" timestamp NULL, "updated_at" timestamp NULL, "deleted" integer(1) NOT NULL DEFAULT 0);
CREATE TABLE "users" ("uuid" varchar(36) primary key NULL, "password" varchar(255) NULL, "pw_func" varchar(255) NULL, "pw_alg" varchar(255) NULL, "pw_cost" integer NULL, "pw_key_size" integer NULL, "pw_nonce" varchar(255) NULL, "email" varchar(255) NULL, "created_at" timestamp NULL, "updated_at" timestamp NULL);
CREATE INDEX user_uuid ON items (user_uuid);
CREATE INDEX user_content on items (user_uuid, content_type);
CREATE INDEX updated_at on items (updated_at);
CREATE INDEX email on users (email);
COMMIT;
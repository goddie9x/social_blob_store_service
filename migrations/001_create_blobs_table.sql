CREATE TYPE target_type_enum AS ENUM (
    'post',
    'comment',
    'user_avatar',
    'user_cover',
    'group_avatar',
    'group_cover',
    'custom_sticker',
    'story',
    'page_avatar',
    'page_cover',
    'chat_cover',
    'chat_avatar'
);

CREATE TYPE blob_type_enum AS ENUM ('video', 'image');

CREATE TABLE IF NOT EXISTS blobs (
    id TEXT PRIMARY KEY,
    filename TEXT NOT NULL,
    size BIGINT NOT NULL,
    content_type TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    last_modified TIMESTAMP NOT NULL,
    data_oid OID NOT NULL,
    owner_id TEXT NOT NULL,
    target_id TEXT NOT NULL,
    target_type target_type_enum NOT NULL,
    type blob_type_enum NOT NULL
);

CREATE INDEX IF NOT EXISTS index_by_target ON blobs(target_id);

CREATE INDEX IF NOT EXISTS index_by_owner_and_target ON blobs(owner_id, target_id);

CREATE INDEX IF NOT EXISTS index_by_owner_and_target_type ON blobs(owner_id, target_type);
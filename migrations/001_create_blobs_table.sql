CREATE TYPE target_type_enum AS ENUM (
    'post', 'comment', 'user_avatar', 'user_cover',
    'group_avatar', 'group_cover', 'custom_sticker',
    'story', 'page_avatar', 'page_cover'
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
    type blob_type_enum NOT NULL,
    CONSTRAINT owner_index UNIQUE(owner_id),
    CONSTRAINT target_type_index UNIQUE(target_id, target_type, type)
);

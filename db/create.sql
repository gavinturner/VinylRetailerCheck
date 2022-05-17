
CREATE DATABASE vinylretailers;

-- connect to this new database
-- \c vinylretailers

CREATE USER vinylretailers WITH NOCREATEDB  ENCRYPTED PASSWORD 'vinylretailers_Cure667';

CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE artists (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    image_url TEXT,
    website_url TEXT
);

CREATE TABLE users_following_artists (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    artist_id BIGINT NOT NULL REFERENCES artists(id)
);

CREATE TABLE retailers (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    url TEXT NOT NULL
);

CREATE TABLE releases (
    id BIGSERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    artist_id BIGINT NOT NULL REFERENCES artists(id),
    UNIQUE (artist_id, title)
);

CREATE TABLE sku (
     id BIGSERIAL PRIMARY KEY,
     retailer_id BIGINT NOT NULL REFERENCES retailers(id),
     release_id BIGINT NOT NULL REFERENCES releases(id),
     artist_id BIGINT NOT NULL REFERENCES artists(id),
     item_url TEXT NOT NULL,
     image_url TEXT,
     price  TEXT NOT NULL
);


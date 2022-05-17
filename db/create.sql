
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
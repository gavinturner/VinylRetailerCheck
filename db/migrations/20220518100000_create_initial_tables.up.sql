
CREATE TABLE IF NOT EXISTS  users (
    id BIGSERIAL PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
GRANT ALL PRIVILEGES ON TABLE users TO vinylretailers;

CREATE TABLE IF NOT EXISTS artists (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    image_url TEXT,
    website_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
GRANT ALL PRIVILEGES ON TABLE artists TO vinylretailers;

CREATE TABLE IF NOT EXISTS users_following_artists (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    artist_id BIGINT NOT NULL REFERENCES artists(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
GRANT ALL PRIVILEGES ON TABLE users_following_artists TO vinylretailers;

CREATE TABLE IF NOT EXISTS retailers (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    url TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
GRANT ALL PRIVILEGES ON TABLE retailers TO vinylretailers;

CREATE TABLE IF NOT EXISTS releases (
    id BIGSERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    artist_id BIGINT NOT NULL REFERENCES artists(id),
    UNIQUE (artist_id, title),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
GRANT ALL PRIVILEGES ON TABLE releases TO vinylretailers;

CREATE TABLE IF NOT EXISTS skus (
    id BIGSERIAL PRIMARY KEY,
    retailer_id BIGINT NOT NULL REFERENCES retailers(id),
    release_id BIGINT NOT NULL REFERENCES releases(id),
    artist_id BIGINT NOT NULL REFERENCES artists(id),
    item_url TEXT NOT NULL,
    image_url TEXT,
    price  TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(release_id, retailer_id, created_at)
);
GRANT ALL PRIVILEGES ON TABLE skus TO vinylretailers;


INSERT INTO retailers (id, name, url) VALUES
    (1, 'Artist First', 'https://artistfirst.com.au'),
    (2, 'Poison City Records', 'https://poisoncityrecords.com')
;
INSERT INTO artists (name, image_url,website_url) VALUES
    ('Clowns', 'https://www.hysteriamag.com/ahm/uploads/2019/02/clowns_hysteria_slider-980x600.jpg', 'https://clownsband.com'),
    ('Hard-ons', 'https://au.rollingstone.com/wp-content/uploads/2021/08/hardons.jpg', 'https://hard-ons1.bandcamp.com'),
    ('Frenzal Rhomb', 'http://pilerats.com/assets/Uploads/frenzal-rhomb-interview-2019.jpg', 'http://www.frenzalrhomb.com.au'),
    ('The Meanies', 'https://images.thebrag.com/td/uploads/2017/07/meanies.jpg', 'https://www.themeanies.net'),
    ('Architects', 'https://cdn.mos.cms.futurecdn.net/VznLMhVnE5U3m4iGUhHNUS-1200-80.jpg', 'https://architectsofficial.com'),
    ('The Neptune Power Federation', 'https://townsquare.media/site/846/files/2019/07/neptune-power.jpg?w=1200&h=0&zc=1&s=0&a=t&q=89', 'http://www.theneptunepowerfederation.com'),
    ('Bad Astronaut', 'https://i0.wp.com/thatsgoodenoughforme.com/wp-content/uploads/2021/03/Bad-Astronaut-band-photo.jpg?fit=479%2C320&ssl=1', 'http://badastronaut.com'),
    ('Lagwagon', 'https://www.theinertia.com/wp-content/uploads/2014/10/Lagwagon_LisaJohnson_sepia.jpg', 'https://www.lagwagon.com'),
    ('Bad Cop Bad Cop', 'https://www.jadedinchicago.com/wp-content/uploads/2015/09/BCBC011.jpg', 'https://www.badcopbadcopband.com/'),
    ('Descendents', 'https://media.altpress.com/uploads/2020/10/Descendents.jpg', 'https://descendents.tumblr.com/'),
    ('Bombpops', 'https://theseeker.ca/wp-content/uploads/2017/02/Bombpops-Fat-Wreck-Chords-Interview.jpg', 'http://thebombpops.com/'),
    ('Pears', 'https://i0.wp.com/wallofsoundau.com/wp-content/uploads/2020/02/pEARS-2559403954-1580781322581.jpg?resize=940%2C400&ssl=1', 'http://pearstheband.com/'),
    ('Lillingtons', '', ''),
    ('Teenage Bottlerocket', '', ''),
    ('NOFX', '', ''),
    ('Propagandhi', '', ''),
    ('PUP', '', ''),
    ('Radiohead', '', ''),
    ('Ramones', '', ''),
    ('Turbonegro', '', ''),
    ('The Good The Bad and The Zugly', '', '')
;


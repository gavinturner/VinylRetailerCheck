-- CREATE DATABASE vinylretailers;
-- connect to this new database
-- \c vinylretailers

CREATE USER vinylretailers WITH NOCREATEDB  ENCRYPTED PASSWORD 'vinylretailers';
GRANT pg_write_all_data TO vinylretailers;


DELETE FROM retailers;
INSERT INTO retailers (id, name, url) VALUES
    (1, 'Artist First', 'https://artistfirst.com.au'),
    (2, 'Poison City Records', 'https://poisoncityrecords.com')
;
SELECT setval('retailers_id_seq', 2, true);

DELETE FROM releases;
DELETE FROM artists;
SELECT setval('artists_id_seq', 1, false);
SELECT setval('releases_id_seq', 1, false);
INSERT INTO artists (name, image_url,website_url) VALUES
    ('Clowns', 'https://www.hysteriamag.com/ahm/uploads/2019/02/clowns_hysteria_slider-980x600.jpg', 'https://clownsband.com'),
    ('Hard-ons', 'https://au.rollingstone.com/wp-content/uploads/2021/08/hardons.jpg', 'https://hard-ons1.bandcamp.com'),
    ('Frenzal Rhomb', 'http://pilerats.com/assets/Uploads/frenzal-rhomb-interview-2019.jpg', 'http://www.frenzalrhomb.com.au'),
    ('The Meanies', 'https://images.thebrag.com/td/uploads/2017/07/meanies.jpg', 'https://www.themeanies.net'),
    ('Architects', 'https://cdn.mos.cms.futurecdn.net/VznLMhVnE5U3m4iGUhHNUS-1200-80.jpg', 'https://architectsofficial.com'),
    ('The Neptune Power Federation', 'https://townsquare.media/site/846/files/2019/07/neptune-power.jpg?w=1200&h=0&zc=1&s=0&a=t&q=89', 'http://www.theneptunepowerfederation.com'),
    ('Bad Astronaught', 'https://i0.wp.com/thatsgoodenoughforme.com/wp-content/uploads/2021/03/Bad-Astronaut-band-photo.jpg?fit=479%2C320&ssl=1', 'http://badastronaut.com'),
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

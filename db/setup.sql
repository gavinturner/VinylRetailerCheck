-- CREATE DATABASE vinylretailers;
-- connect to this new database
-- \c vinylretailers

CREATE USER vinylretailers WITH NOCREATEDB  ENCRYPTED PASSWORD 'vinylretailers';
GRANT pg_write_all_data TO vinylretailers;

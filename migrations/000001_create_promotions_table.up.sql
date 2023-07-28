CREATE TABLE promotions (
    pk  BIGINT PRIMARY KEY,
    "id"  VARCHAR,
    price  VARCHAR,
    expiration_date TIMESTAMP,
    "version" INT
);

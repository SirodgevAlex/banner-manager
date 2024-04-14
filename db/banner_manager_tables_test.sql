CREATE DATABASE test_database;

\c test_database;

CREATE TABLE IF NOT EXISTS test_users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(200) NOT NULL,
    password VARCHAR(200) NOT NULL,
    is_admin BOOLEAN DEFAULT false
);

CREATE TABLE IF NOT EXISTS test_banners (
    id SERIAL PRIMARY KEY,
    feature_id INT,
    tag_id INT,
    title TEXT,
    text TEXT,
    url TEXT,
    is_active BOOLEAN,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

SELECT * FROM test_users;

SELECT * FROM test_banners;

DROP DATABASE IF EXISTS test_database;

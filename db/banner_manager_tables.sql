CREATE TABLE IF NOT EXISTS users (
    ID SERIAL PRIMARY KEY,
    Email VARCHAR(200) NOT NULL,
    Password VARCHAR(200) NOT NULL,
    IsAdmin BOOLEAN DEFAULT false
);

CREATE TABLE banners (
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
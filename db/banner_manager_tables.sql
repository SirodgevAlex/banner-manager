CREATE TABLE IF NOT EXISTS users (
    Id SERIAL PRIMARY KEY,
    Email VARCHAR(200) NOT NULL,
    Password VARCHAR(200) NOT NULL
);

select * from users;
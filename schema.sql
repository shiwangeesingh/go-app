CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    age INT NOT NULL,
    gender TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password BYTEA NOT NULL  -- Fix: Added data type and missing comma
);

CREATE TABLE user_generation (
    id SERIAL PRIMARY KEY,
    user_id INT UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    generation Text NOT NULL, -- Correct ENUM reference
    grade TEXT NOT NULL
);


-- Add an index on the `email` column for faster lookups
CREATE INDEX idx_users_email ON users (email);

-- Add an index on `name` (use LOWER() for case-insensitive searches)
CREATE INDEX idx_users_name ON users (LOWER(name));

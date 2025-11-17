-- Create user_profiles table (only for additional fields not in users)
CREATE TABLE user_profiles (
    id INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    phone VARCHAR(20),
    address TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index for user profiles
CREATE INDEX idx_user_profiles_id ON user_profiles(id);


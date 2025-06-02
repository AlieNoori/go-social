CREATE TABLE IF NOT EXISTS followers(
 user_id bigint REFERENCES users(id) ON DELETE CASCADE NOT NULL,
 follower_id bigint REFERENCES users(id) ON DELETE CASCADE NOT NULL,
 created_at TIMESTAMP(0) with time zone NOT NULL DEFAULT NOW(),
 PRIMARY KEY (user_id, follower_id)
);

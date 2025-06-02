CREATE TABLE IF NOT EXISTS comments (
  id bigserial PRIMARY KEY,
  post_id bigserial references posts(id) ON DELETE CASCADE NOT NULL,
  user_id bigserial references users(id) ON DELETE CASCADE NOT NULL,
  content text NOT NULL,
  created_at TIMESTAMP(0) with time zone NOT NULL DEFAULT NOW()
)

-- +goose Up

CREATE TABLE IF NOT EXISTS user_category_likes (
   user_id     UUID      NOT NULL,
   category    TEXT      NOT NULL,
   cnt         BIGINT    NOT NULL DEFAULT 0,
   created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
   updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
   PRIMARY KEY (user_id, category),
   CONSTRAINT fk_ucl_user
       FOREIGN KEY(user_id)
           REFERENCES users(id)
           ON DELETE CASCADE
);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION trigger_set_user_category_likes_timestamp() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER set_user_category_likes_updated_at
    BEFORE UPDATE ON user_category_likes
    FOR EACH ROW
EXECUTE PROCEDURE trigger_set_user_category_likes_timestamp();

-- +goose Down

DROP TRIGGER IF EXISTS set_user_category_likes_updated_at ON user_category_likes;
DROP FUNCTION IF EXISTS trigger_set_user_category_likes_timestamp();
DROP TABLE IF EXISTS user_category_likes;
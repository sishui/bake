DROP TABLE IF EXISTS test_all_types;

-- enum 类型
DROP TYPE IF EXISTS enum_test_all_types_enum_val;
CREATE TYPE enum_test_all_types_enum_val AS ENUM ('A', 'B', 'C');

CREATE TABLE test_all_types (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

  tiny_int_val SMALLINT,
  small_int_val SMALLINT NOT NULL DEFAULT 0,
  medium_int_val INTEGER,
  int_val INTEGER NOT NULL DEFAULT 0,
  big_int_val BIGINT,

  decimal_val NUMERIC(10,2),
  float_val REAL,
  double_val DOUBLE PRECISION,

  char_val CHAR(10),
  varchar_val VARCHAR(255),
  text_val TEXT,
  mediumtext_val TEXT,
  longtext_val TEXT,

  date_val DATE,
  time_val TIME,
  datetime_val TIMESTAMP,
  timestamp_val TIMESTAMP,

  year_val SMALLINT,

  bool_val BOOLEAN,

  enum_val enum_test_all_types_enum_val,

  set_val TEXT,

  binary_val BYTEA,
  varbinary_val BYTEA,

  blob_val BYTEA,
  mediumblob_val BYTEA,
  longblob_val BYTEA,

  json_val JSONB,

  bit_val BIT(8),

  nullable_int_val INTEGER,

  ignore_val TEXT,

  created_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP(3)
);

COMMENT ON COLUMN "public"."test_all_types"."json_val" IS 'This is a JSON column
with multiple lines of text
This is a JSON column
with multiple lines of text
This is a JSON column
with multiple lines of text';

DROP TABLE IF EXISTS mail;

CREATE TABLE mail (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

  uid BIGINT NOT NULL,

  subject VARCHAR(128) NOT NULL,
  content TEXT NOT NULL,

  status SMALLINT NOT NULL DEFAULT 0,
  has_attachment SMALLINT NOT NULL DEFAULT 0,

  created_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_uid_created ON mail (uid, created_at DESC);
CREATE INDEX idx_uid_status ON mail (uid, status);

DROP TABLE IF EXISTS mail_attachment;

CREATE TABLE mail_attachment (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

  mail_id BIGINT NOT NULL,
  kind SMALLINT NOT NULL DEFAULT 0,

  reward_id BIGINT NOT NULL,
  reward_count INTEGER NOT NULL CHECK (reward_count > 0),
  created_at BIGINT NOT NULL DEFAULT 0,
  updated_at BIGINT NOT NULL
);

CREATE INDEX idx_mail_id ON mail_attachment (mail_id);


INSERT INTO test_all_types (
    tiny_int_val, small_int_val, medium_int_val, int_val, big_int_val,
    decimal_val, float_val, double_val,
    char_val, varchar_val, text_val, mediumtext_val, longtext_val,
    date_val, time_val, datetime_val, timestamp_val, year_val,
    bool_val, enum_val, set_val,
    binary_val, varbinary_val,
    blob_val, mediumblob_val, longblob_val,
    json_val, bit_val,
    nullable_int_val, ignore_val
) VALUES (
    1, 100, 1000, 10000, 100000,
    123.45, 1.23, 3.1415926,
    'char_test', 'varchar_test', 'text...', 'medium text...', 'long text...',
    '2024-01-01', '12:34:56', '2024-01-01 12:34:56', CURRENT_TIMESTAMP, 2024,
    TRUE, 'A', 'X,Y',
    E'\\x61626364656667', E'\\x76617262696e6172795f64617461',  -- BYTEA 示例
    E'\\x626c6f6264617461', E'\\x6d656469756d626c6f6264617461', E'\\x6c6f6e67626c6f6264617461',
    '{"key":"value"}'::jsonb,
    B'10101010',
    NULL, 'ignore me'
);


INSERT INTO mail (uid, subject, content, status, has_attachment)
VALUES (1001, 'Welcome', 'Welcome to the system', 0, 0),
       (1001, 'Reward Mail', 'You got rewards', 1, 1),
       (1002, 'System Notice', 'Maintenance notice', 0, 0);


INSERT INTO mail_attachment (mail_id, kind, reward_id, reward_count, created_at, updated_at)
VALUES
    (2, 1, 1001, 10, EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT),
    (2, 2, 2001, 5, EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT),
    (2, 1, 3001, 1, EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT);

-- 外键示例: users + posts
DROP TABLE IF EXISTS posts;
DROP TABLE IF EXISTS users;

CREATE TABLE users (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  name VARCHAR(64) NOT NULL,
  email VARCHAR(128) NOT NULL UNIQUE,
  created_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE posts (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  user_id BIGINT NOT NULL,
  title VARCHAR(256) NOT NULL,
  content TEXT,
  status SMALLINT NOT NULL DEFAULT 0,
  created_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

  CONSTRAINT fk_posts_user_id FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE INDEX idx_posts_user_id ON posts (user_id);

INSERT INTO users (name, email) VALUES
  ('Alice', 'alice@example.com'),
  ('Bob', 'bob@example.com');

INSERT INTO posts (user_id, title, content, status) VALUES
  (1, 'First Post', 'Hello World!', 1),
  (1, 'Second Post', 'Another post by Alice', 0),
  (2, 'Bob Post', 'Post by Bob', 1);

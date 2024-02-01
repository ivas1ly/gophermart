-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users(
  id uuid PRIMARY KEY,
  username VARCHAR(255) UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,
  current_balance BIGINT NOT NULL CHECK (accrual >= 0) DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
  deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS users_username_idx ON users (username);

CREATE TABLE IF NOT EXISTS orders(
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL,
  number TEXT UNIQUE NOT NULL,
  status VARCHAR(255) NOT NULL,
  accrual BIGINT NOT NULL CHECK (accrual >= 0) DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
  deleted_at TIMESTAMPTZ,
  CONSTRAINT fk_users FOREIGN KEY (user_id) REFERENCES users (id)
);

CREATE INDEX IF NOT EXISTS orders_user_id_idx ON orders (user_id);
CREATE INDEX IF NOT EXISTS orders_number_idx ON orders (number);

CREATE TABLE IF NOT EXISTS withdrawals(
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL,
  order_number TEXT NOT NULL UNIQUE,
  withdrawn BIGINT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
  deleted_at TIMESTAMPTZ,
  CONSTRAINT fk_users FOREIGN KEY (user_id) REFERENCES users (id)
);

CREATE INDEX IF NOT EXISTS withdrawals_user_id_idx ON withdrawals (user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE withdrawals;
DROP TABLE orders;
DROP TABLE users;
-- +goose StatementEnd

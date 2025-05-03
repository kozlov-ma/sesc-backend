-- +goose Up
-- +goose StatementBegin
CREATE TABLE departments (
    id UUID PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    description TEXT
);

CREATE TABLE users (
    id UUID PRIMARY KEY,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    middle_name TEXT DEFAULT '' NOT NULL,
    picture_url TEXT,
    suspended BOOLEAN DEFAULT false,
    department_id UUID REFERENCES departments (id),
    role_id INT,
    auth_id UUID
);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE departments,
users CASCADE;

-- +goose StatementEnd

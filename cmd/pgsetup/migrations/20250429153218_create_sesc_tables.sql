-- +goose Up
-- +goose StatementBegin
CREATE TABLE permissions (
    id UUID PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    description TEXT
);

CREATE TABLE roles (id UUID PRIMARY KEY, name VARCHAR(255) UNIQUE);

CREATE TABLE permissions_roles (
    permission_id UUID REFERENCES permissions (id) ON DELETE CASCADE,
    role_id UUID REFERENCES roles (id) ON DELETE CASCADE,
    PRIMARY KEY (permission_id, role_id)
);

CREATE TABLE departments (
    id UUID PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    head_user_id UUID
);

CREATE TABLE users (
    id UUID PRIMARY KEY,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    middle_name TEXT DEFAULT '' NOT NULL,
    picture_url TEXT,
    suspended BOOLEAN DEFAULT false,
    department_id UUID REFERENCES departments (id),
    role_id UUID REFERENCES roles (id),
    auth_id UUID NOT NULL
);

ALTER TABLE departments ADD FOREIGN KEY (head_user_id) REFERENCES users (id);

CREATE TABLE users_extra_permissions (
    user_id UUID REFERENCES users (id),
    permission_id UUID REFERENCES permissions (id),
    PRIMARY KEY (user_id, permission_id)
);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE permissions_roles,
roles,
permissions,
departments,
users,
users_extra_permissions CASCADE;

-- +goose StatementEnd

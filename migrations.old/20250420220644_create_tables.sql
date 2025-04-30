-- +goose Up
-- +goose StatementBegin
CREATE TYPE permission AS ENUM (
    'admin',
    'create_achlist',
    'verify_dephead',
    'verify_scientific'
);

CREATE TABLE roles (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    permissions permission[] NOT NULL DEFAULT '{}'
);

CREATE TABLE departments (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    head_user_id UUID NOT NULL
);

CREATE TABLE users (
    id UUID PRIMARY KEY,
    last_name TEXT NOT NULL,
    first_name TEXT NOT NULL,
    middle_name TEXT,
    role_id UUID NOT NULL REFERENCES roles (id),
    department_id UUID REFERENCES departments (id),
    photo TEXT NOT NULL,
    suspended BOOLEAN NOT NULL DEFAULT FALSE
);

ALTER TABLE departments ADD FOREIGN KEY (head_user_id) REFERENCES users (id);

CREATE TABLE users_ext (
    user_id UUID PRIMARY KEY REFERENCES users (id),
    job_title TEXT NOT NULL,
    employment_rate TEXT NOT NULL,
    employment_type INTEGER NOT NULL,
    personnel_category INTEGER NOT NULL,
    category TEXT NOT NULL,
    academic_degree INTEGER NOT NULL,
    academic_title TEXT,
    honors JSON,
    date_of_employment DATE NOT NULL
);

CREATE TABLE auth_users (
    user_id UUID PRIMARY KEY REFERENCES users (id),
    login TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL
);

CREATE TYPE achievement_kind AS ENUM ('contest', 'scientific', 'development');

CREATE TABLE document_templates (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL
);

CREATE TABLE achievement_templates (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    reward INTEGER NOT NULL,
    kind achievement_kind NOT NULL
);

CREATE TABLE achievement_templates_documents_templates (
    achievement_template_id UUID NOT NULL REFERENCES achievement_templates (id),
    document_template_id UUID NOT NULL REFERENCES document_templates (id),
    PRIMARY KEY (achievement_template_id, document_template_id)
);

CREATE TYPE achievement_status AS ENUM ('pending', 'approved', 'rejected');

-- Correct order: create list forms and lists before achievements
CREATE TABLE achievement_list_forms (
    id UUID PRIMARY KEY,
    teacher_user_id UUID NOT NULL REFERENCES users (id),
    purpose TEXT
);

CREATE TABLE achievement_lists (
    id UUID PRIMARY KEY,
    teacher_user_id UUID NOT NULL REFERENCES users (id),
    purpose TEXT NOT NULL,
    reviewer_user_id UUID REFERENCES users (id),
    submitted_date DATE NOT NULL
);

CREATE TABLE achievements (
    id UUID PRIMARY KEY,
    list_id UUID REFERENCES achievement_lists (id),
    list_form_id UUID REFERENCES achievement_list_forms (id),
    user_id UUID NOT NULL REFERENCES users (id),
    template_id UUID NOT NULL REFERENCES achievement_templates (id),
    status achievement_status NOT NULL,
    rejected_by_user_id UUID REFERENCES users (id),
    approved_by_user_id UUID REFERENCES users (id),
    comment TEXT
);

CREATE TABLE files (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    owner_user_id UUID NOT NULL REFERENCES users (id),
    created_at TIMESTAMP NOT NULL,
    download_url TEXT NOT NULL,
    size BIGINT NOT NULL,
    file_type TEXT NOT NULL
);

CREATE TABLE documents (
    id UUID PRIMARY KEY,
    file_id UUID REFERENCES files (id),
    achievement_id UUID NOT NULL REFERENCES achievements (id),
    template_id UUID NOT NULL REFERENCES document_templates (id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE
documents,
document_templates,
files,
achievement_lists,
achievement_list_forms,
achievements,
achievement_templates_documents_templates,
achievement_templates,
auth_users,
users_ext,
users,
departments,
roles CASCADE;

DROP TYPE achievement_status, achievement_kind, permission;
-- +goose StatementEnd

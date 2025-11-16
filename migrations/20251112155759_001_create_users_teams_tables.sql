-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id VARCHAR(20) PRIMARY KEY,
    username VARCHAR(100) UNIQUE NOT NULL,
    is_active BOOLEAN DEFAULT true
);

CREATE TABLE teams (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL
);

CREATE TABLE team_members (
    id SERIAL PRIMARY KEY,
    team_id SERIAL NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id VARCHAR(20) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(team_id, user_id)
);

CREATE TABLE pull_requests (
    id VARCHAR(20) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    author_id VARCHAR(20) REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'MERGED')),
    merged_at TIMESTAMP
);

CREATE TABLE pr_reviewers (
    id SERIAL PRIMARY KEY,
    pr_id VARCHAR(20) NOT NULL REFERENCES pull_requests(id) ON DELETE CASCADE,
    user_id VARCHAR(20) NOT NULL REFERENCES users(id),
    UNIQUE(pr_id, user_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS pr_reviewers;

DROP TABLE IF EXISTS pull_requests;

DROP TABLE IF EXISTS team_members;

DROP TABLE IF EXISTS users;

DROP TABLE IF EXISTS teams;
-- +goose StatementEnd

CREATE TABLE user_roles (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) NOT NULL,
    org_id INTEGER REFERENCES organizations(id) NOT NULL,
    org_view BOOLEAN DEFAULT FALSE,
    org_edit BOOLEAN DEFAULT FALSE,
    org_admin BOOLEAN DEFAULT FALSE,
    UNIQUE (user_id, org_id)
);
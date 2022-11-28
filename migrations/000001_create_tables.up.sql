CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(20) NOT NULL,

    -- indices
    UNIQUE (name)
);

CREATE TABLE orgs (
    id SERIAL PRIMARY KEY,
    name text NOT NULL
);

CREATE TABLE org_members (
    id SERIAL PRIMARY KEY,
    org_id integer NOT NULL REFERENCES orgs(id) ON DELETE RESTRICT,
    user_id integer NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    CONSTRAINT org_members_org_id_user_id_key UNIQUE (org_id, user_id)
);

CREATE UNIQUE INDEX orgs_name ON orgs(name);

CREATE TABLE roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(20) NOT NULL,

    -- indices
    UNIQUE (name)
);

CREATE TABLE user_roles (
    -- foreign keys
    user_id integer NOT NULL,
    role_id integer NOT NULL
);

CREATE TABLE batch_changes (
    id serial PRIMARY KEY ,
    name VARCHAR(100) NOT NULL,
    private BOOLEAN DEFAULT false NOT NULL,

    namespace_user_id integer REFERENCES users(id) ON DELETE CASCADE DEFERRABLE,
    namespace_org_id integer REFERENCES orgs(id) ON DELETE CASCADE DEFERRABLE,

    -- foreign keys
    creator_id INTEGER NOT NULL,

    -- indices
    UNIQUE(name, creator_id),

    -- constraints
    CONSTRAINT batch_changes_has_1_namespace CHECK ((namespace_user_id IS NULL) <> (namespace_org_id IS NULL))
);

CREATE TABLE notebooks (
    id serial PRIMARY KEY ,
    name varchar(50) NOT NULL,
    content TEXT NOT NULL,
    private BOOLEAN DEFAULT false NOT NULL,

    -- foreign keys
    creator_id INTEGER NOT NULL ,

    -- indices
    UNIQUE(name, creator_id)
);

CREATE TABLE code_insights (
    id serial PRIMARY KEY ,
    name varchar(50) NOT NULL,

    -- foreign keys
    user_id INTEGER NOT NULL ,

    -- indices
    UNIQUE(name, user_id)
);

CREATE TABLE permissions (
    id SERIAL PRIMARY KEY,
    namespace text NOT NULL,
    action TEXT NOT NULL,
    CONSTRAINT namespace_not_blank CHECK ((namespace <> ''::text))
);


CREATE TABLE role_permissions (
    -- foreign keys
    permission_id integer,
    role_id integer
);

CREATE TABLE batch_changes_namespace (
    id SERIAL PRIMARY KEY,
    resource_id integer NOT NULL REFERENCES batch_changes(id) ON DELETE CASCADE DEFERRABLE,
    subject_id integer NOT NULL REFERENCES users(id) ON DELETE CASCADE DEFERRABLE,
    action VARCHAR(255) NOT NULL
);

CREATE TABLE notebooks_namespace (
    id SERIAL PRIMARY KEY,
    resource_id integer NOT NULL REFERENCES notebooks(id) ON DELETE CASCADE DEFERRABLE,
    subject_id integer NOT NULL REFERENCES users(id) ON DELETE CASCADE DEFERRABLE,
    action VARCHAR(255) NOT NULL
);

CREATE TABLE code_insights_namespace (
    id SERIAL PRIMARY KEY,
    resource_id integer NOT NULL REFERENCES code_insights(id) ON DELETE CASCADE DEFERRABLE,
    subject_id integer NOT NULL REFERENCES users(id) ON DELETE CASCADE DEFERRABLE,
    action VARCHAR(255) NOT NULL
);

ALTER TABLE ONLY batch_changes
    ADD CONSTRAINT batch_change_owner_fkey FOREIGN KEY (creator_id) REFERENCES users(id) ON DELETE CASCADE;

CREATE INDEX batch_changes_private_idx ON batch_changes USING btree (private);

ALTER TABLE ONLY notebooks
    ADD CONSTRAINT notebook_owner_fkey FOREIGN KEY (creator_id) REFERENCES users(id) ON DELETE CASCADE;

CREATE INDEX notebook_private_idx ON notebooks USING btree (private);

ALTER TABLE ONLY code_insights
    ADD CONSTRAINT code_insight_owner_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE ONLY user_roles
    ADD CONSTRAINT user_role_owner_fkey FOREIGN KEY (user_id) REFERENCES users(id);

ALTER TABLE ONLY user_roles
    ADD CONSTRAINT user_role_role_fkey FOREIGN KEY (role_id) REFERENCES roles(id);

ALTER TABLE ONLY user_roles
    ADD CONSTRAINT user_roles_pkey PRIMARY KEY (role_id, user_id);

ALTER TABLE ONLY role_permissions
    ADD CONSTRAINT role_permissions_permission_fkey FOREIGN KEY (permission_id) REFERENCES permissions(id);

ALTER TABLE ONLY role_permissions
    ADD CONSTRAINT role_permission_role_fkey FOREIGN KEY (role_id) REFERENCES roles(id);

ALTER TABLE ONLY role_permissions
    ADD CONSTRAINT role_permissions_pkey PRIMARY KEY (role_id, permission_id);

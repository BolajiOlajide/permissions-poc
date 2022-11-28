-- remove namespaces here
DROP TABLE IF EXISTS batch_changes_namespace;
DROP TABLE IF EXISTS notebooks_namespace;
DROP TABLE IF EXISTS code_insights_namespace;

-- Add tables with exactly one foreign keys here
DROP TABLE IF EXISTS batch_changes;
DROP TABLE IF EXISTS notebooks;
DROP TABLE IF EXISTS code_insights;

-- ADd table with more than one foreign keys
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS org_members;
DROP TABLE IF EXISTS permissions;

-- Add tables without foreign keys here
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS orgs;
DROP TABLE IF EXISTS roles;

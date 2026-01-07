#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-'EOSQL'
  -- Extensions
  CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

  -- Monay schema
  CREATE SCHEMA IF NOT EXISTS monay;

  -- Local roles (script runs only on first init, so no IF NOT EXISTS needed)
  CREATE USER monay_admin     NOINHERIT CREATEROLE LOGIN NOREPLICATION PASSWORD 'postgres';
  CREATE USER monay_readonly  NOINHERIT LOGIN NOREPLICATION PASSWORD 'postgres';
  CREATE USER monay_readwrite NOINHERIT LOGIN NOREPLICATION PASSWORD 'postgres';

  GRANT USAGE ON SCHEMA monay TO monay_readonly, monay_readwrite, monay_admin;
  GRANT ALL PRIVILEGES ON DATABASE monay TO monay_admin;
  ALTER SCHEMA monay OWNER TO monay_admin;

  -- Search paths
  ALTER USER monay_readonly  SET search_path = monay, public;
  ALTER USER monay_readwrite SET search_path = monay, public;
  ALTER USER monay_admin     SET search_path = monay, public;

  -- Grant on future tables
  ALTER DEFAULT PRIVILEGES IN SCHEMA monay GRANT ALL ON TABLES TO monay_admin;
  ALTER DEFAULT PRIVILEGES IN SCHEMA monay GRANT SELECT ON TABLES TO monay_readonly;
  ALTER DEFAULT PRIVILEGES IN SCHEMA monay GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO monay_readwrite;
EOSQL

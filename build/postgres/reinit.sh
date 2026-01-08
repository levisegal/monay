#!/bin/bash
# Idempotent script to reinitialize monay schema and roles.
# Run from host: ./reinit.sh
# Requires psql installed locally
set -e

PGHOST="${PGHOST:-localhost}"
PGPORT="${PGPORT:-6432}"
PGUSER="${PGUSER:-postgres}"
PGPASSWORD="${PGPASSWORD:-postgres}"
PGDATABASE="${PGDATABASE:-monay}"

export PGPASSWORD

psql -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d "$PGDATABASE" -v ON_ERROR_STOP=1 <<-'EOSQL'
  -- Extensions
  CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

  -- Monay schema
  CREATE SCHEMA IF NOT EXISTS monay;

  -- Roles (idempotent with DO block)
  DO $$
  BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'monay_admin') THEN
      CREATE USER monay_admin NOINHERIT CREATEROLE LOGIN NOREPLICATION PASSWORD 'postgres';
    END IF;
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'monay_readonly') THEN
      CREATE USER monay_readonly NOINHERIT LOGIN NOREPLICATION PASSWORD 'postgres';
    END IF;
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'monay_readwrite') THEN
      CREATE USER monay_readwrite NOINHERIT LOGIN NOREPLICATION PASSWORD 'postgres';
    END IF;
  END
  $$;

  -- Grants (idempotent)
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

echo "monay schema and roles initialized"




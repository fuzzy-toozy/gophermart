SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET client_min_messages = warning;
SET row_security = off;

BEGIN;

DO $$
BEGIN
IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'order_status') THEN
        CREATE TYPE ORDER_STATUS AS enum ('NEW', 'PROCESSED', 'PROCESSING', 'INVALID');
END IF;
END$$;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;

SET search_path = public;
SET default_tablespace = '';


CREATE TABLE IF NOT EXISTS users
(
    id uuid NOT NULL DEFAULT uuid_generate_v1mc(),
    username VARCHAR NOT NULL UNIQUE,
    user_password text NOT NULL,
    CONSTRAINT users_pk PRIMARY KEY (id)
);
CREATE UNIQUE INDEX IF NOT EXISTS login_idx
    on users (username);


CREATE TABLE IF NOT EXISTS orders
(
    number      VARCHAR PRIMARY KEY,
    username    VARCHAR NOT NULL REFERENCES users (username),
    uploaded_at TIMESTAMP WITH TIME ZONE NOT NULL,
    status      ORDER_STATUS NOT NULL,
    accrual     FLOAT DEFAULT 0
);
CREATE UNIQUE INDEX IF NOT EXISTS orders_number_idx
    on orders (number);


CREATE TABLE IF NOT EXISTS balances
(
    id           uuid NOT NULL DEFAULT uuid_generate_v1mc(),
    username     VARCHAR NOT NULL REFERENCES users (username),
    processed_at TIMESTAMP WITH TIME ZONE NOT NULL,
    income       FLOAT DEFAULT 0 NOT NULL,
    outcome      FLOAT DEFAULT 0 NOT NULL,
    order_number VARCHAR NOT NULL,
    CONSTRAINT balances_pk PRIMARY KEY (id)
);

COMMIT;
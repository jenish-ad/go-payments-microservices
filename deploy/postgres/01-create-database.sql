-- Runs once, when the Postgres database is first created
-- One database per service: each service owns its data

CREATE database wallet;
CREATE database auth;
CREATE database transfer;

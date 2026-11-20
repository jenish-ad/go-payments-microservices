-- Runs once, when the Postgres database is first created
-- One database per service: each service owns it's data

CREATE database wallet;
CREATE database auth;
CREATE database transfer;

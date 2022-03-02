/* contrib/pgstattuple/pgstattuple--1.4.sql */

-- complain if script is sourced in psql, rather than via CREATE EXTENSION
\echo Use "CREATE EXTENSION pgstattuple" to load this file. \quit

CREATE FUNCTION pgstattuple(IN relname text,
    OUT table_len BIGINT,		-- physical table length in bytes
    OUT tuple_count BIGINT,		-- number of live tuples
    OUT tuple_len BIGINT,		-- total tuples length in bytes
    OUT tuple_percent FLOAT8,		-- live tuples in %
    OUT dead_tuple_count BIGINT,	-- number of dead tuples
    OUT dead_tuple_len BIGINT,		-- total dead tuples length in bytes
    OUT dead_tuple_percent FLOAT8,	-- dead tuples in %
    OUT free_space BIGINT,		-- free space in bytes
    OUT free_percent FLOAT8)		-- free space in %
AS 'MODULE_PATHNAME', 'pgstattuple'
LANGUAGE C STRICT PARALLEL SAFE;

CREATE FUNCTION pgstatindex(IN relname text,
    OUT version INT,
    OUT tree_level INT,
    OUT index_size BIGINT,
    OUT root_block_no BIGINT,
    OUT internal_pages BIGINT,
    OUT leaf_pages BIGINT,
    OUT empty_pages BIGINT,
    OUT deleted_pages BIGINT,
    OUT avg_leaf_density FLOAT8,
    OUT leaf_fragmentation FLOAT8)
AS 'MODULE_PATHNAME', 'pgstatindex'
LANGUAGE C STRICT PARALLEL SAFE;

CREATE FUNCTION pg_relpages(IN relname text)
RETURNS BIGINT
AS 'MODULE_PATHNAME', 'pg_relpages'
LANGUAGE C STRICT PARALLEL SAFE;

/* New stuff in 1.1 begins here */

CREATE FUNCTION pgstatginindex(IN relname regclass,
    OUT version INT4,
    OUT pending_pages INT4,
    OUT pending_tuples BIGINT)
AS 'MODULE_PATHNAME', 'pgstatginindex'
LANGUAGE C STRICT PARALLEL SAFE;

/* New stuff in 1.2 begins here */

CREATE FUNCTION pgstattuple(IN reloid regclass,
    OUT table_len BIGINT,		-- physical table length in bytes
    OUT tuple_count BIGINT,		-- number of live tuples
    OUT tuple_len BIGINT,		-- total tuples length in bytes
    OUT tuple_percent FLOAT8,		-- live tuples in %
    OUT dead_tuple_count BIGINT,	-- number of dead tuples
    OUT dead_tuple_len BIGINT,		-- total dead tuples length in bytes
    OUT dead_tuple_percent FLOAT8,	-- dead tuples in %
    OUT free_space BIGINT,		-- free space in bytes
    OUT free_percent FLOAT8)		-- free space in %
AS 'MODULE_PATHNAME', 'pgstattuplebyid'
LANGUAGE C STRICT PARALLEL SAFE;

CREATE FUNCTION pgstatindex(IN relname regclass,
    OUT version INT,
    OUT tree_level INT,
    OUT index_size BIGINT,
    OUT root_block_no BIGINT,
    OUT internal_pages BIGINT,
    OUT leaf_pages BIGINT,
    OUT empty_pages BIGINT,
    OUT deleted_pages BIGINT,
    OUT avg_leaf_density FLOAT8,
    OUT leaf_fragmentation FLOAT8)
AS 'MODULE_PATHNAME', 'pgstatindexbyid'
LANGUAGE C STRICT PARALLEL SAFE;

CREATE FUNCTION pg_relpages(IN relname regclass)
RETURNS BIGINT
AS 'MODULE_PATHNAME', 'pg_relpagesbyid'
LANGUAGE C STRICT PARALLEL SAFE;

/* New stuff in 1.3 begins here */

CREATE FUNCTION pgstattuple_approx(IN reloid regclass,
    OUT table_len BIGINT,               -- physical table length in bytes
    OUT scanned_percent FLOAT8,         -- what percentage of the table's pages was scanned
    OUT approx_tuple_count BIGINT,      -- estimated number of live tuples
    OUT approx_tuple_len BIGINT,        -- estimated total length in bytes of live tuples
    OUT approx_tuple_percent FLOAT8,    -- live tuples in % (based on estimate)
    OUT dead_tuple_count BIGINT,        -- exact number of dead tuples
    OUT dead_tuple_len BIGINT,          -- exact total length in bytes of dead tuples
    OUT dead_tuple_percent FLOAT8,      -- dead tuples in % (based on estimate)
    OUT approx_free_space BIGINT,       -- estimated free space in bytes
    OUT approx_free_percent FLOAT8)     -- free space in % (based on estimate)
AS 'MODULE_PATHNAME', 'pgstattuple_approx'
LANGUAGE C STRICT PARALLEL SAFE;

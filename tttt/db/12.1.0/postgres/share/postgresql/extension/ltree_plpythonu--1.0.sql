/* contrib/ltree_plpython/ltree_plpythonu--1.0.sql */

-- complain if script is sourced in psql, rather than via CREATE EXTENSION
\echo Use "CREATE EXTENSION ltree_plpythonu" to load this file. \quit

CREATE FUNCTION ltree_to_plpython(val internal) RETURNS internal
LANGUAGE C STRICT IMMUTABLE
AS 'MODULE_PATHNAME';

CREATE TRANSFORM FOR ltree LANGUAGE plpythonu (
    FROM SQL WITH FUNCTION ltree_to_plpython(internal)
);

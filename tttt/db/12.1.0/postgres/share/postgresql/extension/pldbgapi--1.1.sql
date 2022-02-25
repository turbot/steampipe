-- pldbg.sql
--  This script creates the data types and functions defined by the PL debugger API
--
-- Copyright (c) 2004-2018 EnterpriseDB Corporation. All Rights Reserved.
--
-- Licensed under the Artistic License v2.0, see 
--		https://opensource.org/licenses/artistic-license-2.0
-- for full details

\echo Installing pldebugger as unpackaged objects. If you are using PostgreSQL
\echo version 9.1 or above, use "CREATE EXTENSION pldbgapi" instead.

CREATE TYPE breakpoint AS ( func OID, linenumber INTEGER, targetName TEXT );
CREATE TYPE frame      AS ( level INT, targetname TEXT, func OID, linenumber INTEGER, args TEXT );

CREATE TYPE var		   AS ( name TEXT, varClass char, lineNumber INTEGER, isUnique bool, isConst bool, isNotNull bool, dtype OID, value TEXT );
CREATE TYPE proxyInfo  AS ( serverVersionStr TEXT, serverVersionNum INT, proxyAPIVer INT, serverProcessID INT );

CREATE FUNCTION pldbg_oid_debug( functionOID OID ) RETURNS INTEGER AS '$libdir/plugin_debugger' LANGUAGE C STRICT;

-- for backwards-compatibility
CREATE FUNCTION plpgsql_oid_debug( functionOID OID ) RETURNS INTEGER AS $$ SELECT pldbg_oid_debug($1) $$ LANGUAGE sql STRICT;

CREATE FUNCTION pldbg_abort_target( session INTEGER ) RETURNS SETOF boolean AS  '$libdir/plugin_debugger' LANGUAGE C STRICT;
CREATE FUNCTION pldbg_attach_to_port( portNumber INTEGER ) RETURNS INTEGER AS '$libdir/plugin_debugger' LANGUAGE C STRICT;
CREATE FUNCTION pldbg_continue( session INTEGER ) RETURNS breakpoint AS '$libdir/plugin_debugger' LANGUAGE C STRICT;
CREATE FUNCTION pldbg_create_listener() RETURNS INTEGER AS '$libdir/plugin_debugger' LANGUAGE C STRICT;
CREATE FUNCTION pldbg_deposit_value( session INTEGER, varName TEXT, lineNumber INTEGER, value TEXT ) RETURNS boolean AS  '$libdir/plugin_debugger' LANGUAGE C STRICT;
CREATE FUNCTION pldbg_drop_breakpoint( session INTEGER, func OID, linenumber INTEGER ) RETURNS boolean AS  '$libdir/plugin_debugger' LANGUAGE C STRICT;
CREATE FUNCTION pldbg_get_breakpoints( session INTEGER ) RETURNS SETOF breakpoint AS '$libdir/plugin_debugger' LANGUAGE C STRICT;
CREATE FUNCTION pldbg_get_source( session INTEGER, func OID ) RETURNS TEXT AS '$libdir/plugin_debugger' LANGUAGE C STRICT;
CREATE FUNCTION pldbg_get_stack( session INTEGER ) RETURNS SETOF frame AS '$libdir/plugin_debugger' LANGUAGE C STRICT;
CREATE FUNCTION pldbg_get_proxy_info( ) RETURNS proxyInfo AS '$libdir/plugin_debugger' LANGUAGE C STRICT;
CREATE FUNCTION pldbg_get_variables( session INTEGER ) RETURNS SETOF var AS '$libdir/plugin_debugger' LANGUAGE C STRICT;
CREATE FUNCTION pldbg_select_frame( session INTEGER, frame INTEGER ) RETURNS breakpoint AS '$libdir/plugin_debugger' LANGUAGE C STRICT;
CREATE FUNCTION pldbg_set_breakpoint( session INTEGER, func OID, linenumber INTEGER ) RETURNS boolean AS  '$libdir/plugin_debugger' LANGUAGE C STRICT;
CREATE FUNCTION pldbg_set_global_breakpoint( session INTEGER, func OID, linenumber INTEGER, targetPID INTEGER ) RETURNS boolean AS  '$libdir/plugin_debugger' LANGUAGE C;
CREATE FUNCTION pldbg_step_into( session INTEGER ) RETURNS breakpoint AS '$libdir/plugin_debugger' LANGUAGE C STRICT;
CREATE FUNCTION pldbg_step_over( session INTEGER ) RETURNS breakpoint AS '$libdir/plugin_debugger' LANGUAGE C STRICT;
CREATE FUNCTION pldbg_wait_for_breakpoint( session INTEGER ) RETURNS breakpoint  AS '$libdir/plugin_debugger' LANGUAGE C STRICT;
CREATE FUNCTION pldbg_wait_for_target( session INTEGER ) RETURNS INTEGER AS '$libdir/plugin_debugger' LANGUAGE C STRICT;

/*
 * pldbg_get_target_info() function can be used to return information about
 * a function.
 *
 * Deprecated. This is used by the pgAdmin debugger GUI, but new applications
 * should just query the catalogs directly.
 */
CREATE TYPE targetinfo AS ( target OID, schema OID, nargs INT, argTypes oidvector, targetName NAME, argModes "char"[], argNames TEXT[], targetLang OID, fqName TEXT, returnsSet BOOL, returnType OID,

  -- The following columns are only needed when running in an EnterpriseDB
  -- server. On PostgreSQL, we return just dummy values for them.
  --
  -- 'isFunc' and 'pkg' only make sense on EnterpriseDB.  'isfunc' is true
  -- if the function is a regular function, not a stored procedure or a
  -- function that was created implictly to back a trigger created with the
  -- Oracle-compatible CREATE TRIGGER syntax. If the function belongs to a
  -- package, 'pkg' is the package's OID, or 0 otherwise.
  --
  -- 'argDefVals' is a representation of the function's argument DEFAULTs.
  -- That would be nice to have on PostgreSQL as well. Unfortunately our
  -- current implementation relies on an EDB-only function to get that
  -- information, so we cannot just use it as is. TODO: rewrite that using
  -- pg_get_expr(pg_proc.proargdefaults).
  isFunc BOOL,
  pkg OID,
  argDefVals TEXT[]
);

-- Create the pldbg_get_target_info() function. We use an inline code block
-- so that we can check and create it slightly differently if running on
-- an EnterpriseDB server.

DO $do$

declare
  isedb bool;
  createstmt text;
begin

  isedb = (SELECT version() LIKE '%EnterpriseDB%');

  createstmt := $create_stmt$

CREATE FUNCTION pldbg_get_target_info(signature text, targetType "char") returns targetinfo AS $$
  SELECT p.oid AS target,
         pronamespace AS schema,
         pronargs::int4 AS nargs,
         -- The returned argtypes column is of type oidvector, but unlike
         -- proargtypes, it's supposed to include OUT params. So we
         -- essentially have to return proallargtypes, converted to an
         -- oidvector. There is no oid[] -> oidvector cast, so we have to
         -- do it via text.
         CASE WHEN proallargtypes IS NOT NULL THEN
           translate(proallargtypes::text, ',{}', ' ')::oidvector
         ELSE
           proargtypes
         END AS argtypes,
         proname AS targetname,
         proargmodes AS argmodes,
         proargnames AS proargnames,
         prolang AS targetlang,
         quote_ident(nspname) || '.' || quote_ident(proname) AS fqname,
         proretset AS returnsset,
         prorettype AS returntype,
$create_stmt$;

-- Add the three EDB-columns to the query (as dummies if we're installing
-- to PostgreSQL)
IF isedb THEN
  createstmt := createstmt ||
$create_stmt$
         p.protype='0' AS isfunc,
         CASE WHEN n.nspparent <> 0 THEN n.oid ELSE 0 END AS pkg,
	 edb_get_func_defvals(p.oid) AS argdefvals
$create_stmt$;
ELSE
  createstmt := createstmt ||
$create_stmt$
         't'::bool AS isfunc,
         0::oid AS pkg,
	 NULL::text[] AS argdefvals
$create_stmt$;
END IF;
  -- End of conditional part

  createstmt := createstmt ||
$create_stmt$
  FROM pg_proc p, pg_namespace n
  WHERE p.pronamespace = n.oid
  AND p.oid = $1::oid
  -- We used to support querying by function name or trigger name/oid as well,
  -- but that was never used in the client, so the support for that has been
  -- removed. The targeType argument remains as a legacy of that. You're
  -- expected to pass 'o' as target type, but it doesn't do anything.
  AND $2 = 'o'
$$ LANGUAGE SQL;
$create_stmt$;

  execute createstmt;

-- Add a couple of EDB specific functions
IF isedb THEN
   CREATE FUNCTION edb_oid_debug(functionOID oid) RETURNS integer AS $$
     select pldbg_oid_debug($1);
   $$ LANGUAGE SQL;

   CREATE FUNCTION pldbg_get_pkg_cons(packageOID oid) RETURNS oid AS $$
     select oid from pg_proc where pronamespace=$1 and proname='cons';
   $$ LANGUAGE SQL;
END IF;

end;
$do$;

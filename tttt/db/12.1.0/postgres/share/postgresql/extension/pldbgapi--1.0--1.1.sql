DO $do$

declare
  isedb bool;
  createstmt text;
begin

  isedb = (SELECT version() LIKE '%EnterpriseDB%');

  createstmt := $create_stmt$

CREATE OR REPLACE FUNCTION pldbg_get_target_info(signature text, targetType "char") returns targetinfo AS $$
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
   CREATE OR REPLACE FUNCTION edb_oid_debug(functionOID oid) RETURNS integer AS $$
     select pldbg_oid_debug($1);
   $$ LANGUAGE SQL;

   CREATE OR REPLACE FUNCTION pldbg_get_pkg_cons(packageOID oid) RETURNS oid AS $$
     select oid from pg_proc where pronamespace=$1 and proname='cons';
   $$ LANGUAGE SQL;
END IF;

end;
$do$;

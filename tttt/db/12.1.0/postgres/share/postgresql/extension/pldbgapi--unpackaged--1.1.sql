ALTER EXTENSION pldbgapi ADD TYPE breakpoint;
ALTER EXTENSION pldbgapi ADD TYPE frame;

ALTER EXTENSION pldbgapi ADD TYPE targetinfo;

ALTER EXTENSION pldbgapi ADD TYPE var;
ALTER EXTENSION pldbgapi ADD TYPE proxyInfo;

ALTER EXTENSION pldbgapi ADD FUNCTION plpgsql_oid_debug( functionOID OID );

ALTER EXTENSION pldbgapi ADD FUNCTION pldbg_abort_target( session INTEGER );
ALTER EXTENSION pldbgapi ADD FUNCTION pldbg_attach_to_port( portNumber INTEGER );
ALTER EXTENSION pldbgapi ADD FUNCTION pldbg_continue( session INTEGER );
ALTER EXTENSION pldbgapi ADD FUNCTION pldbg_create_listener();
ALTER EXTENSION pldbgapi ADD FUNCTION pldbg_deposit_value( session INTEGER, varName TEXT, lineNumber INTEGER, value TEXT );
ALTER EXTENSION pldbgapi ADD FUNCTION pldbg_drop_breakpoint( session INTEGER, func OID, linenumber INTEGER );
ALTER EXTENSION pldbgapi ADD FUNCTION pldbg_get_breakpoints( session INTEGER );
ALTER EXTENSION pldbgapi ADD FUNCTION pldbg_get_source( session INTEGER, func OID );
ALTER EXTENSION pldbgapi ADD FUNCTION pldbg_get_stack( session INTEGER );
ALTER EXTENSION pldbgapi ADD FUNCTION pldbg_get_proxy_info( );
ALTER EXTENSION pldbgapi ADD FUNCTION pldbg_get_variables( session INTEGER );
ALTER EXTENSION pldbgapi ADD FUNCTION pldbg_select_frame( session INTEGER, frame INTEGER );
ALTER EXTENSION pldbgapi ADD FUNCTION pldbg_set_breakpoint( session INTEGER, func OID, linenumber INTEGER );
ALTER EXTENSION pldbgapi ADD FUNCTION pldbg_set_global_breakpoint( session INTEGER, func OID, linenumber INTEGER, targetPID INTEGER );
ALTER EXTENSION pldbgapi ADD FUNCTION pldbg_step_into( session INTEGER );
ALTER EXTENSION pldbgapi ADD FUNCTION pldbg_step_over( session INTEGER );
ALTER EXTENSION pldbgapi ADD FUNCTION pldbg_wait_for_breakpoint( session INTEGER );
ALTER EXTENSION pldbgapi ADD FUNCTION pldbg_wait_for_target( session INTEGER );
ALTER EXTENSION pldbgapi ADD FUNCTION pldbg_get_target_info( signature TEXT, targetType "char" );

DO $do$

declare
  isedb bool;
begin

  isedb = (SELECT version() LIKE '%EnterpriseDB%');

  -- Add a couple of EDB specific functions
  IF isedb THEN
    ALTER EXTENSION pldbgapi ADD edb_oid_debug( functionOID oid );
    ALTER EXTENSION pldbgapi ADD pldbg_get_pkg_cons( packageOID oid );
  END IF;

$do$;

package db_local

const cloneForeignSchemaSQL = `CREATE OR REPLACE FUNCTION clone_foreign_schema(
    source_schema text,
    dest_schema text,
    plugin_name text)
    RETURNS text AS
$BODY$

DECLARE	
    src_oid          oid;
    object           text;
    dest_table       text;
    table_sql        text;
    columns_sql      text;
    type_            text;
    column_          text;
    underlying_type  text;
    res              text;
BEGIN

    -- Check that source_schema exists
    SELECT oid INTO src_oid
    FROM pg_namespace
    WHERE nspname = source_schema;
    IF NOT FOUND
    THEN
        RAISE EXCEPTION 'source schema % does not exist!', source_schema;
        RETURN '';
    END IF;

    -- Create schema
    EXECUTE 'DROP SCHEMA IF EXISTS "' ||  dest_schema || '" CASCADE';
    EXECUTE 'CREATE SCHEMA "' || dest_schema || '"';
    EXECUTE 'GRANT USAGE ON SCHEMA "' || dest_schema || '" TO steampipe_users';
    EXECUTE 'ALTER DEFAULT PRIVILEGES IN SCHEMA "' || dest_schema || '" GRANT SELECT ON TABLES TO steampipe_users';

    -- Create tables
    FOR object IN
        SELECT TABLE_NAME::text
        FROM information_schema.tables
        WHERE table_schema = source_schema
          AND table_type = 'FOREIGN'
    LOOP
        columns_sql := '';

        FOR column_, type_ IN
            SELECT c.column_name::text, 
                   CASE 
                       WHEN c.data_type = 'USER-DEFINED' THEN t.typname
                       ELSE c.data_type
                   END as data_type
            FROM information_schema.COLUMNS c
            LEFT JOIN pg_catalog.pg_type t ON c.udt_name = t.typname
            WHERE c.table_schema = source_schema
              AND c.TABLE_NAME = object
        LOOP
            IF columns_sql <> ''
            THEN
                columns_sql = columns_sql || ',';
            END IF;
            columns_sql = columns_sql || quote_ident(column_) || ' ' || type_;
        END LOOP;

        dest_table := '"' || dest_schema || '".' || quote_ident(object);
        table_sql :='CREATE FOREIGN TABLE ' || dest_table || ' (' || columns_sql || ') SERVER steampipe OPTIONS (table '|| $$'$$ || quote_ident(object) || $$'$$ || ') ';
        EXECUTE table_sql;

        SELECT CONCAT(res, table_sql, ';') into res;
    END LOOP;
    RETURN res;
END

$BODY$
LANGUAGE plpgsql VOLATILE
                 COST 100;
`

const cloneCommentsSQL = `
CREATE OR REPLACE FUNCTION clone_table_comments(
    source_schema text,
    dest_schema text)
    RETURNS text AS
$BODY$

DECLARE
    src_oid         oid;
    dest_oid        oid;
    t               text;
    ret             text;
    query           text;
    table_desc      text;
    column_desc     text;
    column_number   int;
    c               text;
BEGIN

    -- Check that source_schema and dest_schema exist
    SELECT oid INTO src_oid
    FROM pg_namespace
    WHERE nspname = quote_ident(source_schema);
    IF NOT FOUND
    THEN
        RAISE NOTICE 'source schema % does not exist!', source_schema;
        RETURN 'source schema does not exist!';
    END IF;

    SELECT oid INTO dest_oid
    FROM pg_namespace
    WHERE nspname = quote_ident(dest_schema);
    IF NOT FOUND
    THEN
        RAISE NOTICE 'dest schema % does not exist!', dest_schema;
        RETURN 'dest schema does not exist!';
    END IF;


    -- Copy comments
    FOR t IN
        SELECT table_name::text
        FROM information_schema.tables
            WHERE table_schema = quote_ident(source_schema)
            AND table_type = 'FOREIGN'
    LOOP
        SELECT OBJ_DESCRIPTION((quote_ident(source_schema) || '.' || quote_ident(t))::REGCLASS) INTO table_desc;
        query = 'COMMENT ON FOREIGN TABLE ' || quote_ident(dest_schema) ||  '.' || quote_ident(t) || ' IS $steampipe_escape$' || table_desc || '$steampipe_escape$';
       SELECT CONCAT(ret, query || '\n') INTO ret;
        EXECUTE query;

        FOR  c,column_number IN
            SELECT column_name, ordinal_position
            FROM information_schema.COLUMNS
                WHERE table_schema = quote_ident(source_schema)
                AND table_name = quote_ident(t)
        LOOP
            SELECT PG_CATALOG.COL_DESCRIPTION((quote_ident(source_schema) || '.' || quote_ident(t))::REGCLASS::OID, column_number) INTO column_desc;
            query = 'COMMENT ON COLUMN ' || quote_ident(dest_schema) ||  '.' || quote_ident(t) ||  '.' || quote_ident(c) || ' IS $steampipe_escape$' || column_desc || '$steampipe_escape$';
--            SELECT CONCAT(ret, query || '\n') INTO ret;
            EXECUTE query;
        END LOOP;
    END LOOP;

    RETURN ret;
END

$BODY$
    LANGUAGE plpgsql VOLATILE
                     COST 100;
`

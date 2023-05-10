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
    table_sql      text;
    columns_sql      text;
    type_            text;
    column_          text;
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
    EXECUTE 'DROP SCHEMA IF EXISTS ' ||  dest_schema || ' CASCADE';
    EXECUTE 'CREATE SCHEMA ' || dest_schema;
--     EXECUTE 'COMMENT ON SCHEMA ' || dest_schema || 'IS  ''steampipe plugin: ' || plugin_name || '''';
    EXECUTE 'GRANT USAGE ON SCHEMA ' || dest_schema || ' TO steampipe_users';
    EXECUTE 'ALTER DEFAULT PRIVILEGES IN SCHEMA ' || dest_schema || ' GRANT SELECT ON TABLES TO steampipe_users';

-- Create tables
    FOR object IN
        SELECT TABLE_NAME::text
        FROM information_schema.tables
        WHERE table_schema = source_schema
          AND table_type = 'FOREIGN'

        LOOP
            columns_sql := '';

            FOR column_, type_ IN
                SELECT column_name::text, data_type::text
                FROM information_schema.COLUMNS
                WHERE table_schema = source_schema
                  AND TABLE_NAME = object

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

const cloneCommentsSQL = `CREATE OR REPLACE FUNCTION clone_foreign_schema(
    source_schema text,
    dest_schema text,
    plugin_name text)
    RETURNS text AS
$BODY$

DECLARE
    src_oid          oid;
    object           text;
    dest_table       text;
    table_sql      text;
    columns_sql      text;
    type_            text;
    column_          text;
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
    EXECUTE 'DROP SCHEMA IF EXISTS ' ||  dest_schema || ' CASCADE';
    EXECUTE 'CREATE SCHEMA ' || dest_schema;
--     EXECUTE 'COMMENT ON SCHEMA ' || dest_schema || 'IS  ''steampipe plugin: ' || plugin_name || '''';
    EXECUTE 'GRANT USAGE ON SCHEMA ' || dest_schema || ' TO steampipe_users';
    EXECUTE 'ALTER DEFAULT PRIVILEGES IN SCHEMA ' || dest_schema || ' GRANT SELECT ON TABLES TO steampipe_users';

-- Create tables
    FOR object IN
        SELECT TABLE_NAME::text
        FROM information_schema.tables
        WHERE table_schema = source_schema
          AND table_type = 'FOREIGN'

        LOOP
            columns_sql := '';

            FOR column_, type_ IN
                SELECT column_name::text, data_type::text
                FROM information_schema.COLUMNS
                WHERE table_schema = source_schema
                  AND TABLE_NAME = object

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

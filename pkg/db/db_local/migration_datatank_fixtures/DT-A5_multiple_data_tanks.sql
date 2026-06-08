-- DT-A5: Three data tanks coexisting in the same cluster. Each is its own
-- <handle> + <handle>-parts schema pair. The migration must dump/restore ALL
-- data-tank schema pairs, not just one - this fixture asserts the multi-tank
-- enumeration. Tanks: fast_aws, fast_azure, fast_gcp.

create schema if not exists "fast_aws";
create schema if not exists "fast_aws-parts";
create schema if not exists "fast_azure";
create schema if not exists "fast_azure-parts";
create schema if not exists "fast_gcp";
create schema if not exists "fast_gcp-parts";

do $$
declare
    tank text;
    p int;
    pname text;
    parts_schema text;
begin
    foreach tank in array array['fast_aws', 'fast_azure', 'fast_gcp'] loop
        parts_schema := tank || '-parts';
        execute format(
            'create table %I.%I (
                 id bigint, title text, _cloud_partition text, _ctx jsonb,
                 constraint %I primary key (id, _cloud_partition)
             ) partition by list (_cloud_partition)',
            tank, 'resource', tank || '_resource_pk');
        for p in 1..3 loop
            pname := 'part_conn_' || p || '-20260101000000';
            execute format(
                'create table %I.%I (like %I.%I including all)',
                parts_schema, pname, tank, 'resource');
            execute format(
                'alter table %I.%I attach partition %I.%I for values in (%L)',
                tank, 'resource', parts_schema, pname, pname);
            execute format(
                'insert into %I.%I (id, title, _cloud_partition, _ctx)
                 select (%s-1)*10 + g, ''r-'' || ((%s-1)*10 + g), %L,
                        ''{}''::jsonb
                 from generate_series(1, 10) g',
                tank, 'resource', p, p, pname);
        end loop;
    end loop;
end $$;

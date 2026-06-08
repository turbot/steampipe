-- DT-D base fixture for the refresh-mid-flight cases (DT-D1, DT-D2, DT-D3).
-- An ordinary single-tank, multi-partition data tank. The mid-flight events
-- (refresh fires during dump, partition swap in flight) are injected by the
-- harness setup hooks, not by the fixture SQL.

create schema if not exists "fast_aws";
create schema if not exists "fast_aws-parts";

create table "fast_aws"."aws_resource" (
    id bigint,
    title text,
    _cloud_partition text,
    _ctx jsonb,
    constraint aws_resource_pk primary key (id, _cloud_partition)
) partition by list (_cloud_partition);

do $$
declare
    p int;
    pname text;
begin
    for p in 1..3 loop
        pname := 'part_conn_' || p || '-20260101000000';
        execute format(
            'create table %I.%I (like %I.%I including all)',
            'fast_aws-parts', pname, 'fast_aws', 'aws_resource');
        execute format(
            'alter table %I.%I attach partition %I.%I for values in (%L)',
            'fast_aws', 'aws_resource', 'fast_aws-parts', pname, pname);
        execute format(
            'insert into %I.%I (id, title, _cloud_partition, _ctx)
             select (%s-1)*20 + g, ''r-'' || ((%s-1)*20 + g), %L, ''{}''::jsonb
             from generate_series(1, 20) g',
            'fast_aws', 'aws_resource', p, p, pname);
    end loop;
end $$;

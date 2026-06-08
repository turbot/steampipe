-- DT-A2: Single data tank, one table, 4 partitions (100 rows total).
--
-- The task file originally proposed HASH partitioning here. exec-5
-- (data-tank-storage-patterns.md, "Partitioning") establishes that real data
-- tank uses ONLY declarative PARTITION BY LIST(_cloud_partition) - "No
-- PARTITION BY HASH or PARTITION BY RANGE anywhere." DT-A2 is therefore a
-- LIST-partition fixture with 4 children (4 connections / aggregator parts),
-- not HASH. This is the deviation called out in the task header ("if a
-- category below turns out to not apply to real data tank, drop it").

create schema if not exists "fast_aws";
create schema if not exists "fast_aws-parts";

create table "fast_aws"."aws_account" (
    id bigint,
    title text,
    _cloud_partition text,
    _ctx jsonb,
    constraint aws_account_pk primary key (id, _cloud_partition)
) partition by list (_cloud_partition);

do $$
declare
    p int;
    pname text;
begin
    for p in 1..4 loop
        pname := 'part_conn_' || p || '-20260101000000';
        execute format(
            'create table %I.%I (like %I.%I including all)',
            'fast_aws-parts', pname, 'fast_aws', 'aws_account');
        execute format(
            'alter table %I.%I attach partition %I.%I for values in (%L)',
            'fast_aws', 'aws_account', 'fast_aws-parts', pname, pname);
        execute format(
            'insert into %I.%I (id, title, _cloud_partition, _ctx)
             select (%s-1)*25 + g, ''account-'' || ((%s-1)*25 + g), %L,
                    jsonb_build_object(''cloud'', jsonb_build_object(''updated_at'', ''2026-01-01T00:00:00Z''))
             from generate_series(1, 25) g',
            'fast_aws', 'aws_account', p, p, pname);
    end loop;
end $$;

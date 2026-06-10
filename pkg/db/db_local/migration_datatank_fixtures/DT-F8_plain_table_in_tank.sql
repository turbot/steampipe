-- DT-F8: a tank schema containing BOTH a normal partitioned tank table AND a
-- plain (non-partitioned) table with rows.
--
-- The plain table is a legitimate production state, not an invented one: a
-- refresh swap DETACHES the old partition (instantly an ordinary standalone
-- table in the -parts schema) and a separate cleanup workflow drops detached
-- partitions older than 24h - so detached-but-not-yet-dropped plain tables
-- linger in tank schemas routinely. An interrupted refresh similarly orphans
-- its pre-attach (plain) partition table. The copy tiers must carry these
-- tables' rows directly: a plain table's rows live in the table itself, not
-- in partitions.

create schema if not exists "fast_aws";
create schema if not exists "fast_aws-parts";

-- The normal partitioned tank table (2 partitions, 50 rows).
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
    for p in 1..2 loop
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

-- The plain table: a detached old partition awaiting cleanup, sitting in the
-- -parts schema as an ordinary table with 40 rows.
create table "fast_aws-parts"."part_conn_1-20251231000000" (
    id bigint,
    title text,
    _cloud_partition text,
    _ctx jsonb
);
insert into "fast_aws-parts"."part_conn_1-20251231000000" (id, title, _cloud_partition, _ctx)
select g, 'old-account-' || g, 'part_conn_1-20251231000000',
       jsonb_build_object('cloud', jsonb_build_object('updated_at', '2025-12-31T00:00:00Z'))
from generate_series(1, 40) g;

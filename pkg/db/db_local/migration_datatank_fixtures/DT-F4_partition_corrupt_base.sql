-- DT-F4 base: a data tank with many partitions. The harness corrupts the data
-- of ONE partition at restore time. DESIRED behaviour the engine implements:
-- tiers 1-3 fail at the bad partition, tier 4 (per-partition COPY) lands the
-- good partitions but the bad one never migrates - a PARTIAL result, NOT a
-- success. Outcome: dtOutcomePartialRestoreDataPreserved (committed=false; old
-- PG14 data dir + safety dump preserved; deletion gate does not fire).
--
-- 20 partitions keep the fixture small while still exercising "199/200-style"
-- degraded restore (the bad partition is partition 7, injected by the harness).

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
    for p in 1..20 loop
        pname := 'part_conn_' || lpad(p::text, 3, '0') || '-20260101000000';
        execute format(
            'create table %I.%I (like %I.%I including all)',
            'fast_aws-parts', pname, 'fast_aws', 'aws_resource');
        execute format(
            'alter table %I.%I attach partition %I.%I for values in (%L)',
            'fast_aws', 'aws_resource', 'fast_aws-parts', pname, pname);
        execute format(
            'insert into %I.%I (id, title, _cloud_partition, _ctx)
             select (%s-1)*5 + g, ''r-'' || ((%s-1)*5 + g), %L, ''{}''::jsonb
             from generate_series(1, 5) g',
            'fast_aws', 'aws_resource', p, p, pname);
    end loop;
end $$;

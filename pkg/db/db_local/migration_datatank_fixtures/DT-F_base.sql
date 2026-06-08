-- DT-F base fixture: a normal multi-partition data tank used by the tier
-- escalation cases DT-F1..DT-F5. The tier failures (parallel-dependency
-- ordering trip, single-transaction mid-restore error, corrupt partition,
-- all-tier failure) are injected by the harness setup hooks - they cannot be
-- expressed as loadable SQL because they model restore-time / disk-level
-- failures, not source-schema content. DT-F1 (no injection) is the Tier-1
-- success control.

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
    for p in 1..4 loop
        pname := 'part_conn_' || p || '-20260101000000';
        execute format(
            'create table %I.%I (like %I.%I including all)',
            'fast_aws-parts', pname, 'fast_aws', 'aws_resource');
        execute format(
            'alter table %I.%I attach partition %I.%I for values in (%L)',
            'fast_aws', 'aws_resource', 'fast_aws-parts', pname, pname);
        execute format(
            'insert into %I.%I (id, title, _cloud_partition, _ctx)
             select (%s-1)*25 + g, ''r-'' || ((%s-1)*25 + g), %L, ''{}''::jsonb
             from generate_series(1, 25) g',
            'fast_aws', 'aws_resource', p, p, pname);
    end loop;
end $$;

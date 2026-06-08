-- DT-A3: Single data tank, one table, 10 partitions keyed by connection-ID-like
-- value, small. This is the canonical real shape: "Partition per connection"
-- (data-tank-storage-patterns.md / Victor) - a workspace with 10 connections
-- becomes 10 _cloud_partition values. Partition names follow the real format
-- <datatank_part_id>-<YYYYMMDDHHMMSS> (partition.go generatePartitionName).

create schema if not exists "fast_aws";
create schema if not exists "fast_aws-parts";

create table "fast_aws"."aws_resource" (
    id bigint,
    arn text,
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
    for p in 1..10 loop
        -- connection-ID-like part id (20-char cuid-ish), per real partition naming.
        pname := 'part_cuh3t49br37t5n4udc' || lpad(p::text, 2, '0') || '-20260105120000';
        execute format(
            'create table %I.%I (like %I.%I including all)',
            'fast_aws-parts', pname, 'fast_aws', 'aws_resource');
        execute format(
            'alter table %I.%I attach partition %I.%I for values in (%L)',
            'fast_aws', 'aws_resource', 'fast_aws-parts', pname, pname);
        execute format(
            'insert into %I.%I (id, arn, title, _cloud_partition, _ctx)
             select (%s-1)*10 + g, ''arn:aws:'' || ((%s-1)*10 + g), ''res-'' || ((%s-1)*10 + g), %L,
                    jsonb_build_object(''cloud'', jsonb_build_object(''updated_at'', ''2026-01-05T12:00:00Z''))
             from generate_series(1, 10) g',
            'fast_aws', 'aws_resource', p, p, p, pname);
    end loop;
end $$;

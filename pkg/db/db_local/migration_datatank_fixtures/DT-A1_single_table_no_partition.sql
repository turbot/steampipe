-- DT-A1: Single data tank, one table, single partition (100 rows).
--
-- Real data tank ALWAYS uses PARTITION BY LIST(_cloud_partition) even for the
-- "no partition" / aggregator case - the parent is still partitioned, with a
-- single child partition keyed by the aggregator/connection name
-- (even an aggregator data tank with one part
-- still uses PARTITION BY LIST with a single child). There is no
-- non-partitioned standalone-table code path. So "no partitions" here means
-- "one partition", matching the real shape.
--
-- Schemas: target = "fast_aws", parts = "fast_aws-parts" (the hyphen is
-- intrinsic to the -parts naming and forces quoting everywhere).

create schema if not exists "fast_aws";
create schema if not exists "fast_aws-parts";

-- Parent table in the target schema, partitioned by list on the synthetic key.
create table "fast_aws"."aws_account" (
    id bigint,
    title text,
    _cloud_partition text,
    _ctx jsonb,
    constraint aws_account_pk primary key (id, _cloud_partition)
) partition by list (_cloud_partition);

-- One partition table in the -parts schema.
create table "fast_aws-parts"."part_conn_a-20260101000000" (
    like "fast_aws"."aws_account" including all
);

alter table "fast_aws"."aws_account"
    attach partition "fast_aws-parts"."part_conn_a-20260101000000"
    for values in ('part_conn_a-20260101000000');

insert into "fast_aws"."aws_account" (id, title, _cloud_partition, _ctx)
select
    g,
    'account-' || g,
    'part_conn_a-20260101000000',
    jsonb_build_object('cloud', jsonb_build_object('updated_at', '2026-01-01T00:00:00Z'))
from generate_series(1, 100) g;

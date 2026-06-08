-- DT-P18.3 (data-tank shape of catalog item P18.3): an UNLOGGED partitioned data
-- tank. PG14 accepts CREATE UNLOGGED TABLE ... PARTITION BY LIST; PG18 rejects an
-- unlogged partitioned parent at restore time ("partitioned tables cannot be
-- unlogged"). Tiers 1-2 (pg_restore of the dump's unlogged DDL) hit that error, so
-- the tiered engine must escalate to the per-table COPY tier, which recreates the
-- parent as an ordinary LOGGED partitioned table (data tank never relies on the
-- unlogged property) and lands the data. DESIRED outcome: the data migrates via a
-- fallback tier (tier 3) - NOT a hard failure. This is the data-shape counterpart
-- of the public F01 case, where the public engine has no per-table COPY fallback
-- and the same DDL is a graceful restore failure instead.
create schema if not exists "fast_aws";
create schema if not exists "fast_aws-parts";

create unlogged table "fast_aws"."aws_resource" (
    id bigint,
    title text,
    _cloud_partition text,
    _ctx jsonb,
    constraint aws_resource_pk primary key (id, _cloud_partition)
) partition by list (_cloud_partition);

create unlogged table "fast_aws-parts"."part_conn_a-20260101000000" (
    like "fast_aws"."aws_resource" including all
);

alter table "fast_aws"."aws_resource"
    attach partition "fast_aws-parts"."part_conn_a-20260101000000"
    for values in ('part_conn_a-20260101000000');

insert into "fast_aws"."aws_resource" (id, title, _cloud_partition, _ctx)
select g, 'r-' || g, 'part_conn_a-20260101000000', '{}'::jsonb
from generate_series(1, 20) g;

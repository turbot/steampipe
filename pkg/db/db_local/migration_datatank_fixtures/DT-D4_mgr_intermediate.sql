-- DT-D4: a user-driven intra-version schema migration is in flight when the
-- PG-version migration begins. The Pipes data-tank code (cleanup.go, relname like '%_mgr_%')
-- documents the _mgr_ infix convention for "migrating" parent tables that sit
-- under a temporary name until the user-driven swap completes.
--
-- DESIRED behaviour the engine must implement: detect _mgr_ infix tables and
-- either skip them or wait for the cleanup workflow before the PG migration
-- starts. This fixture seeds both a normal data-tank table AND a _mgr_
-- intermediate so the harness can assert the chosen path is honoured.

create schema if not exists "fast_aws";
create schema if not exists "fast_aws-parts";

-- normal parent.
create table "fast_aws"."aws_resource" (
    id bigint,
    title text,
    _cloud_partition text,
    _ctx jsonb,
    constraint aws_resource_pk primary key (id, _cloud_partition)
) partition by list (_cloud_partition);

create table "fast_aws-parts"."part_conn_a-20260101000000" (
    like "fast_aws"."aws_resource" including all
);
alter table "fast_aws"."aws_resource"
    attach partition "fast_aws-parts"."part_conn_a-20260101000000"
    for values in ('part_conn_a-20260101000000');
insert into "fast_aws"."aws_resource" (id, title, _cloud_partition, _ctx)
select g, 'r-' || g, 'part_conn_a-20260101000000', '{}'::jsonb
from generate_series(1, 30) g;

-- in-flight _mgr_ intermediate parent (new column shape), not yet swapped in.
create table "fast_aws"."aws_resource_mgr_20260105120000" (
    id bigint,
    title text,
    extra_col text,
    _cloud_partition text,
    _ctx jsonb,
    constraint aws_resource_mgr_pk primary key (id, _cloud_partition)
) partition by list (_cloud_partition);

-- DT-A4: Aggregator-only data tank. In the real data tank, the aggregator type is NOT a
-- separate storage shape - it is the same parent + parts + PARTITION BY LIST
-- shape with ONE partition keyed by the aggregator's name (rather than per
-- connection). aggregator.go is pure orchestration; storage goes through the
-- same partition.go path. So this fixture is one parent + one partition, named
-- after the aggregator rather than a connection.

create schema if not exists "all_aws";
create schema if not exists "all_aws-parts";

create table "all_aws"."aws_account_agg" (
    id bigint,
    title text,
    _cloud_partition text,
    _ctx jsonb,
    constraint aws_account_agg_pk primary key (id, _cloud_partition)
) partition by list (_cloud_partition);

create table "all_aws-parts"."part_aggregator_all_aws-20260101000000" (
    like "all_aws"."aws_account_agg" including all
);

alter table "all_aws"."aws_account_agg"
    attach partition "all_aws-parts"."part_aggregator_all_aws-20260101000000"
    for values in ('part_aggregator_all_aws-20260101000000');

insert into "all_aws"."aws_account_agg" (id, title, _cloud_partition, _ctx)
select
    g,
    'agg-account-' || g,
    'part_aggregator_all_aws-20260101000000',
    jsonb_build_object('cloud', jsonb_build_object('updated_at', '2026-01-01T00:00:00Z'))
from generate_series(1, 50) g;

-- DT-C1: reserved-word trip wire. A source-plugin column named system_user.
-- PG16+ reserves SYSTEM_USER; a PG14 pg_dump that emits the column unquoted
-- restore-fails on PG18 with a syntax error (catalog item F05). The DESIRED
-- behaviour exec-5b must implement: a reserved-word pre-flight scan flags the
-- column and routes the migration directly to tier 3 (per-table COPY with
-- quoted identifiers), which succeeds. Outcome:
-- dtOutcomeAutoRestoreSucceededAtTier3.
--
-- The column name is the trip wire - everything else is an ordinary data tank.

create schema if not exists "fast_aws";
create schema if not exists "fast_aws-parts";

create table "fast_aws"."aws_iam_event" (
    id bigint,
    system_user text,
    title text,
    _cloud_partition text,
    _ctx jsonb,
    constraint aws_iam_event_pk primary key (id, _cloud_partition)
) partition by list (_cloud_partition);

create table "fast_aws-parts"."part_conn_a-20260101000000" (
    like "fast_aws"."aws_iam_event" including all
);

alter table "fast_aws"."aws_iam_event"
    attach partition "fast_aws-parts"."part_conn_a-20260101000000"
    for values in ('part_conn_a-20260101000000');

insert into "fast_aws"."aws_iam_event" (id, system_user, title, _cloud_partition, _ctx)
select
    g,
    'user-' || g,
    'event-' || g,
    'part_conn_a-20260101000000',
    '{}'::jsonb
from generate_series(1, 20) g;

-- DT-P18.4 (data-tank shape of catalog item P18.4): a data tank whose text column
-- holds a value containing the literal sequence "\." . On PG18, CSV/text COPY no
-- longer treats "\." as end-of-data, but the data-tank migration dumps in directory
-- format and streams partition data with binary COPY, so the value round-trips
-- byte-for-byte. DESIRED outcome: clean migrate at tier 1 - the engine is not
-- tripped by the CSV EOF change, and the "\." row survives.
create schema if not exists "fast_aws";
create schema if not exists "fast_aws-parts";

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

insert into "fast_aws"."aws_resource" (id, title, _cloud_partition, _ctx) values
    (1, 'before', 'part_conn_a-20260101000000', '{}'::jsonb),
    (2, E'a\\.b', 'part_conn_a-20260101000000', '{}'::jsonb),
    (3, 'after', 'part_conn_a-20260101000000', '{}'::jsonb);

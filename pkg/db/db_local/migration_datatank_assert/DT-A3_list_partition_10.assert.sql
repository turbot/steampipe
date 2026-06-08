-- DT-A3 assertion: 10 attached partitions, 100 rows total, valid _ctx.
select
    (select count(*) from "fast_aws"."aws_resource") as row_count,
    (select count(distinct _cloud_partition) from "fast_aws"."aws_resource") as partition_values,
    (select count(*) from pg_partition_tree('"fast_aws"."aws_resource"'::regclass)
       where isleaf) as attached_partitions,
    (select count(*) filter (where _ctx is not null and jsonb_typeof(_ctx) = 'object')
       from "fast_aws"."aws_resource") as valid_ctx;

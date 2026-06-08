-- DT-A2 assertion: 4 attached partitions, 100 rows total, valid _ctx.
select
    (select count(*) from "fast_aws"."aws_account") as row_count,
    (select count(distinct _cloud_partition) from "fast_aws"."aws_account") as partition_values,
    (select count(*) from pg_partition_tree('"fast_aws"."aws_account"'::regclass)
       where isleaf) as attached_partitions;

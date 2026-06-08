-- DT-F base / DT-F1 assertion: 4 attached partitions, 100 rows.
select
    (select count(*) from "fast_aws"."aws_resource") as row_count,
    (select count(distinct _cloud_partition) from "fast_aws"."aws_resource") as partition_values,
    (select count(*) from pg_partition_tree('"fast_aws"."aws_resource"'::regclass)
       where isleaf) as attached_partitions;

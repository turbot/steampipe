-- DT-D base assertion: 3 attached partitions, 60 rows.
select
    (select count(*) from "fast_aws"."aws_resource") as row_count,
    (select count(distinct _cloud_partition) from "fast_aws"."aws_resource") as partition_values,
    (select count(*) from pg_partition_tree('"fast_aws"."aws_resource"'::regclass)
       where isleaf) as attached_partitions;

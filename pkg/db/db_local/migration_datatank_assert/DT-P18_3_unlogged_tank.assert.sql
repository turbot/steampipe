-- DT-P18.3 assertion: 1 attached partition, 20 rows, regardless of the tank being
-- recreated LOGGED on the target. Row count + partition topology must round-trip.
select
    (select count(*) from "fast_aws"."aws_resource") as row_count,
    (select count(distinct _cloud_partition) from "fast_aws"."aws_resource") as partition_values,
    (select count(*) from pg_partition_tree('"fast_aws"."aws_resource"'::regclass)
       where isleaf) as attached_partitions;

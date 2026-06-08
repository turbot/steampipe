-- DT-A4 assertion: aggregator tank, single attached partition, 50 rows.
select
    (select count(*) from "all_aws"."aws_account_agg") as row_count,
    (select count(distinct _cloud_partition) from "all_aws"."aws_account_agg") as partition_values,
    (select count(*) from pg_partition_tree('"all_aws"."aws_account_agg"'::regclass)
       where isleaf) as attached_partitions;

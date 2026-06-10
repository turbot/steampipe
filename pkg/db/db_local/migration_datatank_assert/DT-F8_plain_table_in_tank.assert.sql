-- DT-F8 assertion: the partitioned table's rows AND the plain (detached old
-- partition) table's rows both arrive. Run on both clusters and compared, so
-- a silently-dropped plain table (zero rows / missing relation on the target)
-- diverges from the source.
select
    (select count(*) from "fast_aws"."aws_account") as tank_rows,
    (select count(*) from pg_partition_tree('"fast_aws"."aws_account"'::regclass)
       where isleaf) as attached_partitions,
    (select count(*) from "fast_aws-parts"."part_conn_1-20251231000000") as plain_table_rows,
    (select min(title) from "fast_aws-parts"."part_conn_1-20251231000000") as plain_first_title;

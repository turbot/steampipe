-- DT-D4 assertion: the normal parent survived with its partition AND the _mgr_
-- intermediate parent survived too (skip-or-wait path must not drop it).
select
    (select count(*) from "fast_aws"."aws_resource") as row_count,
    (select count(*) from pg_partition_tree('"fast_aws"."aws_resource"'::regclass)
       where isleaf) as attached_partitions,
    (select count(*) from pg_class c
       join pg_namespace n on n.oid = c.relnamespace
      where n.nspname = 'fast_aws' and c.relname like '%\_mgr\_%') as mgr_tables_present;

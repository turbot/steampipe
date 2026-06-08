-- DT-A5 assertion: all three data tanks survived with 3 attached partitions and
-- 30 rows each. One row per tank; the harness checksums the whole result set.
select tank,
       (select count(*) from pg_partition_tree(format('%I.%I', tank, 'resource')::regclass) where isleaf) as attached_partitions
from (values ('fast_aws'), ('fast_azure'), ('fast_gcp')) as t(tank)
union all
select 'fast_aws_rows', (select count(*) from "fast_aws"."resource")
union all
select 'fast_azure_rows', (select count(*) from "fast_azure"."resource")
union all
select 'fast_gcp_rows', (select count(*) from "fast_gcp"."resource")
order by 1;

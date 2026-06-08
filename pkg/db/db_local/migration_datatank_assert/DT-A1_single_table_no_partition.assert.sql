-- DT-A1 assertion: row count, partition attachment, and JSON validity of _ctx.
-- Run identically against PG14 (golden) and PG18 (migrated); digests must match.
--
-- The parent SELECT only returns rows from ATTACHED children, so row_count and
-- partition_values already imply attachment. attached_partitions cross-checks
-- the partition topology via pg_partition_tree (PG12+, present on both majors).
select
    (select count(*) from "fast_aws"."aws_account") as row_count,
    (select count(*) filter (where _ctx is not null and jsonb_typeof(_ctx) = 'object')
       from "fast_aws"."aws_account") as valid_ctx,
    (select count(distinct _cloud_partition) from "fast_aws"."aws_account") as partition_values,
    (select count(*) from pg_partition_tree('"fast_aws"."aws_account"'::regclass)
       where isleaf) as attached_partitions;

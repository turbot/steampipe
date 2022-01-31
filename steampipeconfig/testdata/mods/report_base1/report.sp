query aws_a3_unencrypted_and_nonversioned_buckets_by_region {
  sql = <<EOF
with unencrypted_buckets_by_region as (
  select
    region,
    count(*) as unencrypted
  from
    aws_morales_aaa.aws_s3_bucket
  where
    server_side_encryption_configuration is null
  group by
    region
),
nonversioned_buckets_by_region as (
  select
    region,
    count(*) as nonversioned
  from
    aws_morales_aaa.aws_s3_bucket
  where
    not versioning_enabled
  group by
    region
),
compliant_buckets_by_region as (
  select
    region,
    count(*) as "other"
  from
    aws_morales_aaa.aws_s3_bucket
  where
    server_side_encryption_configuration is not null
    and versioning_enabled
  group by
    region
)
select
  c.region as "Region",
  coalesce(c.other, 0) as "Compliant",
  coalesce(u.unencrypted, 0) as "Unencrypted",
  coalesce(v.nonversioned, 0) as "Non-Versioned"
from
  compliant_buckets_by_region c
  full join unencrypted_buckets_by_region u on c.region = u.region
  full join nonversioned_buckets_by_region v on c.region = v.region;
EOF
}

chart aws_bucket_info {
  type = "column"
  sql = query.aws_a3_unencrypted_and_nonversioned_buckets_by_region.sql
  grouping = "compare"
  legend {
    position = "bottom"
  }
  axes {
    x {
      title {
        display = "always"
        value = "Foo"
      }
    }
    y {
      title {
        display = "always"
        value = "Foo"
      }
    }
  }
}

report debug {
  title = "Debug"

  chart {
    base = chart.aws_bucket_info
    width = 8
    axes {
      x {
        title {
          value = "Barz"
        }
      }
    }
  }
}
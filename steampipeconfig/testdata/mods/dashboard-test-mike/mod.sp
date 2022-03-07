mod "dashboard_poc" {
  title = "Dashboard POC"
}

#dashboard "my_dashboard" {
#  title = "Dashboard Test"
#  text {
#    value = "# Foo"
#  }
#
#  chart {
#    base = hackernews.chart.companies
#  }
#}


query "aws_s3_buckets_by_versioning_enabled" {
  sql = <<-EOQ
    with versioning as (
      select
        case when versioning_enabled then 'Enabled' else 'Disabled' end as versioning_status,
        region
      from
        aws_s3_bucket
    )
    select
      versioning_status,
      count(versioning_status) as "Total"
    from
      versioning
    where
      region = $1
    group by
      versioning_status
EOQ
  param "region" { }
}



dashboard "inputs" {
  title = "Inputs Test"

  container {
    chart {
      type  = "donut"
      width = 3
      query = query.aws_s3_buckets_by_versioning_enabled
    }
  }
}


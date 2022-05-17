
query "aws_region_input" {
  sql = <<EOQ
select
  title as label,
  region as value
from
  aws_region
where
  account_id = '876515858155'
order by
  title;
EOQ
}

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
  param "region" {
    default = "us-east-1"
  }
}


input "global1" {
  sql = query.aws_region_input.sql
  width = 3
}

dashboard "inputs" {
  title = "Inputs Test"


  input "region" {
    sql = query.aws_region_input.sql
    width = 3
  }

  chart {
    type = "donut"
    width = 5
    query = query.aws_s3_buckets_by_versioning_enabled
    args = {
      "region" = self.input.region.value
    }
    title = "AWS IAM Users MFA Status"

    series "Total" {
      point "Disabled" {
        color = "red"
      }

      point "Enabled" {
        color = "green"
      }
    }
  }
}

dashboard "inputs2" {
  title = "Inputs Test 2"

  input "region" {
    sql = query.aws_region_input.sql
    width = 3
  }


  chart {
    type = "donut"
    width = 5
    query = query.aws_s3_buckets_by_versioning_enabled
    args = {
      "region" = self.input.region.value
    }
    title = "AWS IAM Users MFA Status"

    series "Total" {
      point "Disabled" {
        color = "red"
      }

      point "Enabled" {
        color = "green"
      }
    }
  }
}



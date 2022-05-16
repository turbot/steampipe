
query "aws_iam_users_by_mfa_enabled" {
  sql = <<-EOQ
    with mfa as (
      select
        case when mfa_enabled then 'Enabled' else 'Disabled' end as mfa_status
      from
        aws_iam_user
    )
    select
      mfa_status,
      count(mfa_status) as "Total"
    from
      mfa
    group by
      mfa_status
  EOQ
}

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

dashboard "inputs" {
  title = "Inputs Test"

  text {
    value = "dasboard input"
  }
  input "region" {
    sql   = query.aws_region_input.sql
    width = 3
  }

  container {
    container {
      container {
        text {
          value = "container input"
        }
        input "region2" {
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

        chart {
          type = "pie"
          width = 3
          query = query.aws_s3_buckets_by_versioning_enabled
          args = {
            "region" = self.input.region2.value
          }
          title = "AWS IAM Users MFA Status"
        }
      }
    }
  }
}
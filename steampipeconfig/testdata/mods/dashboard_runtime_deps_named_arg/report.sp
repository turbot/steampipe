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

dashboard dashboard_named_args {
  title = "dashboard with named arguments"

  input "user" {
    title = "AWS IAM User"
    sql   = query.aws_region_input.sql
    width = 4
  }

  table {
    sql = "select 1"
    args  = {
      "iam_user_arn" = self.input.user.value
    }

    column "depth" {
      display = "none"
    }
  }
}
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


dashboard "dashboard_named_args" {
  title = "dashboard with named arguments"

  input "user" {
    title = "AWS IAM User"
    sql   = query.aws_region_input.sql
    width = 4
  }

  table {
    sql = "select * from aws_account where arn in ($1)"
    with "w1" {
        sql = "select * from aws_account"

    }
    args  = {
      "with_val" = with.w1.rows[*].arn
    }
    param "with_val" {}


    column "depth" {
      display = "none"
    }
  }
}
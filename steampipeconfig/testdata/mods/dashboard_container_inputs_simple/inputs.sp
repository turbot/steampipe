
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

dashboard "inputs" {
  title = "Inputs Test"

  container {
    input "region" {
      sql = query.aws_region_input.sql
      width = 3
    }

  }
}
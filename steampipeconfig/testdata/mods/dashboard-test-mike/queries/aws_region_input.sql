select
  title as label,
  region as value
from
  aws_region
where
  account_id = '876515858155'
order by
  title;
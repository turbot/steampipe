select
  title as label,
  vpc_id as value
from
  aws_vpc
where
  account_id = '876515858155'
  and region = 'us-east-1'
order by
  title;
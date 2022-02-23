select
  title as label,
  group_id as value
from
  aws_vpc_security_group
where
  vpc_id = 'vpc-9d7ae1e7'
order by
  title;
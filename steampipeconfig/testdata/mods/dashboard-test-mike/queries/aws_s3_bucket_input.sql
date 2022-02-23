select
  title as label,
  arn as value
from
  aws_s3_bucket
order by
  title;
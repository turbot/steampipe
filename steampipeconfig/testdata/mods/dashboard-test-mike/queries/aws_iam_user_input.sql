select
  title as label,
  arn as value
from
  aws_iam_user
order by
  title;
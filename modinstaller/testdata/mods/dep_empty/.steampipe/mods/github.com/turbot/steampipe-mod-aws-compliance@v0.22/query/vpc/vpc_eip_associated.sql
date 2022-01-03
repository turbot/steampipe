select
  -- Required Columns
  'arn:' || partition || ':ec2:' || region || ':' || account_id || ':eip/' || allocation_id as resource,
  case
    when association_id is null then 'alarm'
    else 'ok'
  end status,
  case
    when association_id is null then title || ' is not associated with any resource.'
    else title || ' is associated with a resource.'
  end reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_vpc_eip;
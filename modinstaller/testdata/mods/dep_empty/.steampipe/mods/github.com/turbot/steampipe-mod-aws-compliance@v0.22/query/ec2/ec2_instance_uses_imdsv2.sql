select
  -- Required Columns
  arn as resource,
  case
    when metadata_options ->> 'HttpTokens' = 'optional' then 'alarm'
    else 'ok'
  end as status,
  case
    when metadata_options ->> 'HttpTokens' = 'optional' then title || ' not configured to use Instance Metadata Service Version 2 (IMDSv2).'
    else title || ' configured to use Instance Metadata Service Version 2 (IMDSv2).'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_ec2_instance;
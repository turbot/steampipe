select
  -- Required Columns
  arn as resource,
  case
    when kms_key_id is null then 'alarm'
    else 'ok'
  end as status,
  case
    when kms_key_id is null then title || ' encryption at rest disabled.'
    else title || ' encryption at rest enabled.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_sagemaker_endpoint_configuration;
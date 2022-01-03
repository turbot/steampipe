select
  -- Required Columns
  arn as resource,
  case
    when client_certificate_id is null then 'alarm'
    else 'ok'
  end as status,
  case
    when client_certificate_id is null then title || ' does not use SSL certificate.'
    else title || ' uses SSL certificate.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_api_gateway_stage;
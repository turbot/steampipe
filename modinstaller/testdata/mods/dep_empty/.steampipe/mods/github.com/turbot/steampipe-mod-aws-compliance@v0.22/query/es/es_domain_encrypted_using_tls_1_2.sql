select
  -- Required Columns
  arn as resource,
  case
    when domain_endpoint_options ->> 'TLSSecurityPolicy' = 'Policy-Min-TLS-1-2-2019-07' then 'ok'
    else 'alarm'
  end status,
  case
    when domain_endpoint_options ->> 'TLSSecurityPolicy' = 'Policy-Min-TLS-1-2-2019-07' then title || ' encrypted using TLS 1.2.'
    else title || ' not encrypted using TLS 1.2.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_elasticsearch_domain;
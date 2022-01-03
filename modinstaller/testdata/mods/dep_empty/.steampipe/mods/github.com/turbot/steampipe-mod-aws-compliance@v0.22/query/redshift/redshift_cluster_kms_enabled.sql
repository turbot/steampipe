select
  -- Required Columns
  'arn:aws:redshift:' || region || ':' || account_id || ':' || 'cluster' || ':' || cluster_identifier as resource,
  case
    when encrypted and kms_key_id is not null then 'ok'
    else 'alarm'
  end as status,
  case
    when encrypted and kms_key_id is not null then title || ' encrypted with KMS.'
    else title || ' not encrypted with KMS'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_redshift_cluster;
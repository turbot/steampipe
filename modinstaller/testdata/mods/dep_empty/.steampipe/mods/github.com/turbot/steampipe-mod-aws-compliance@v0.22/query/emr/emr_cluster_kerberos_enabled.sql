select
  -- Required Columns
  cluster_arn as resource,
  case
    when kerberos_attributes is null then 'alarm'
    else 'ok'
  end as status,
  case
    when kerberos_attributes is null then title || ' Kerberos not enabled.'
    else title || ' Kerberos enabled.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_emr_cluster;
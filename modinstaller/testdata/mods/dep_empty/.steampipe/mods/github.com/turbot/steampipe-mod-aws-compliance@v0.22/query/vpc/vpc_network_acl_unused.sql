select
  -- Required Columns
  network_acl_id as resource,
  case
    when jsonb_array_length(associations) >= 1  then 'ok'
    else 'alarm'
  end status,
  case
    when jsonb_array_length(associations) >= 1 then title || ' associated with subnet.'
    else title || ' not associated with subnet.'
  end reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_vpc_network_acl;
-- This also addresses, Lambda in VPC.
-- As Lambda creates an elastic network interface for each subnet in your function's VPC configuration.
with associated_sg as (
  select
    sg ->> 'GroupId' as secgrp_id,
    sg ->> 'GroupName' as secgrp_name
  from
    aws_ec2_network_interface,
    jsonb_array_elements(groups) as sg
)
select
  -- Required Columns
  distinct s.arn as resource,
  case
    when a.secgrp_id = s.group_id then 'ok'
    else 'alarm'
  end as status,
  case
    when a.secgrp_id = s.group_id then s.title || ' is associated.'
    else s.title || ' not associated.'
  end as reason,
  -- Additional Dimensions
  s.region,
  s.account_id
from
  aws_vpc_security_group s
  left join associated_sg a on s.group_id = a.secgrp_id;

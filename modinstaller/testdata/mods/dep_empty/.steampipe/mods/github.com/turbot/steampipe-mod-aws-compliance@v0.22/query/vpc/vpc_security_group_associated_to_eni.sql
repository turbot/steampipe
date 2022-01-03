with associated_sg as (
  select
    count(sg ->> 'GroupId'),
    sg ->> 'GroupId' as secgrp_id
  from
    aws_ec2_network_interface,
    jsonb_array_elements(groups) as sg
    group by sg ->> 'GroupId'
)
select
  -- Required Columns
  distinct s.arn as resource,
  case
    when a.secgrp_id = s.group_id then 'ok'
    else 'alarm'
  end as status,
  case
    when a.secgrp_id = s.group_id then s.title || ' is associated with ' || a.count || ' ENI(s).'
    else s.title || ' not associated to any ENI.'
  end as reason,
  -- Additional Dimensions
  s.region,
  s.account_id
from
  aws_vpc_security_group as s
  left join associated_sg as a on s.group_id = a.secgrp_id;
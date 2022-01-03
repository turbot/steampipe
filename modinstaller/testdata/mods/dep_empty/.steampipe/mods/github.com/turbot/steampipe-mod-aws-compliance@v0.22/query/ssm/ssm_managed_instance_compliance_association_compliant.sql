select
  -- Required Columns
  id as resource,
  case
    when c.status = 'COMPLIANT' then 'ok'
    else 'alarm'
  end as status,
  case
    when c.status = 'COMPLIANT' then c.resource_id || ' association ' || c.title || ' is compliant.'
    else c.resource_id || ' association ' || c.title || ' is non-compliant.'
  end as reason,
  -- Additional Dimensions
  c.region,
  c.account_id
from
  aws_ssm_managed_instance as i,
  aws_ssm_managed_instance_compliance as c
where
  c.resource_id = i.instance_id
  and c.compliance_type = 'Association';
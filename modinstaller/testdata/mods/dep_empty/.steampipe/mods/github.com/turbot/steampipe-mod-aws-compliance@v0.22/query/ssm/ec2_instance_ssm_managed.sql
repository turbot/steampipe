select
  -- Required Columns
  i.arn as resource,
  case
    when m.instance_id is null then 'alarm'
    else 'ok'
  end as status,
  case
    when m.instance_id is null then i.title || ' not managed by AWS SSM.'
    else i.title || ' managed by AWS SSM.'
  end as reason,
  -- Additional Dimentions
  i.region,
  i.account_id
from
  aws_ec2_instance i
  left join aws_ssm_managed_instance m on m.instance_id = i.instance_id;

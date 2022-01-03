select
  -- Required Columns
  arn as resource,
  case
    when state != 'in-use' then 'skip'
    when attachment ->> 'DeleteOnTermination' = 'true' then 'ok'
    else 'alarm'
  end as status,
  case
    when state != 'in-use' then title || ' not attached to EC2 instance.'
    when attachment ->> 'DeleteOnTermination' = 'true' then title || ' attached to ' || (attachment ->> 'InstanceId') || ', delete on termination enabled.'
    else title || ' attached to ' || (attachment ->> 'InstanceId') || ', delete on termination disabled.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_ebs_volume
  left join jsonb_array_elements(attachments) as attachment on true;

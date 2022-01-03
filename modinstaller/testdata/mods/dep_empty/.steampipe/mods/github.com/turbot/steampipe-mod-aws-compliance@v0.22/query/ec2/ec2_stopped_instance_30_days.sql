select
  -- Required Columns
  arn as resource,
  case
    when instance_state not in ('stopped', 'stopping') then 'skip'
    when state_transition_time <= (current_date - interval '30' day) then 'alarm'
    else 'ok'
  end as status,
  case
    when instance_state not in ('stopped', 'stopping') then title || ' is in ' || instance_state || ' state.'
    else title || ' stopped since ' || to_char(state_transition_time , 'DD-Mon-YYYY') || ' (' || extract(day from current_timestamp - state_transition_time) || ' days).'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_ec2_instance;

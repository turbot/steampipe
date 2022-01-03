with all_plans as (
  select
    arn,
    r as Rules,
    title,
    region,
    account_id
  from
    aws_backup_plan,
    jsonb_array_elements(backup_plan -> 'Rules') as r
)
select
  -- Required Columns
  -- The resource ARN can be duplicate as we are checking all the associated rules to the backup-plan
  -- Backup plans are composed of one or more backup rules.
  -- https://docs.aws.amazon.com/aws-backup/latest/devguide/creating-a-backup-plan.html
  r.arn as resource,
  case
    when r.Rules is null then 'alarm'
    when r.Rules ->> 'Lifecycle' is null then 'ok'
    when (r.Rules -> 'Lifecycle' ->> 'DeleteAfterDays')::int >= 37 then 'ok'
    else 'alarm'
  end as status,
  case
    when r.Rules is null then r.title || ' retention period not set.'
    when r.Rules ->> 'Lifecycle' is null then (r.Rules ->> 'RuleName') || ' retention period set to never expire.'
    else (r.Rules ->> 'RuleName') || ' retention period set to ' || (r.Rules -> 'Lifecycle' ->> 'DeleteAfterDays') || ' days.'
  end as reason,
  -- Additional Dimensions
  r.region,
  r.account_id
from
  all_plans as r;
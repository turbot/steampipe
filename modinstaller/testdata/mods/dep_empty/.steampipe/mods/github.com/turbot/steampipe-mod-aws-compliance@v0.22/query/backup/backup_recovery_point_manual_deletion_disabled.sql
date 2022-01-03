with recovery_point_manual_deletion_disabled as (
  select
    arn
  from
    aws_backup_vault,
    jsonb_array_elements(policy -> 'Statement') as s
  where
    s ->> 'Effect' = 'Deny' and
    s -> 'Action' @> '["backup:DeleteRecoveryPoint","backup:UpdateRecoveryPointLifecycle","backup:PutBackupVaultAccessPolicy"]'
    and s ->> 'Resource' = '*'
  group by
    arn
)
select
  -- Required Columns
  v.arn as resource,
  case
    when d.arn is not null then 'ok'
    else 'alarm'
  end as status,
  case
    when d.arn is not null then v.title || ' recovery point manual deletion disabled.'
    else v.title || ' recovery point manual deletion not disabled.'
  end as reason,
  -- Additional Dimensions
  v.region,
  v.account_id
from
  aws_backup_vault as v
  left join recovery_point_manual_deletion_disabled as d on v.arn = d.arn;
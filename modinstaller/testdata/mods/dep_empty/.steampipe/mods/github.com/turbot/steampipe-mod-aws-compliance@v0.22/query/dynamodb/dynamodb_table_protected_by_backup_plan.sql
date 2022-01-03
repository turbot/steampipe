with backup_protected_table as (
  select
    resource_arn as arn
  from
    aws_backup_protected_resource as b
  where
    resource_type = 'DynamoDB'
)
select
  -- Required Columns
  t.arn as resource,
  case
    when b.arn is not null then 'ok'
    else 'alarm'
  end as status,
  case
    when b.arn is not null then t.title || ' is protected by backup plan.'
    else t.title || ' is not protected by backup plan.'
  end as reason,
  -- Additional Dimensions
  t.region,
  t.account_id
from
  aws_dynamodb_table as t
  left join backup_protected_table as b on t.arn = b.arn;

select
  -- Required Columns
  arn as resource,
  engine,
  case
    when engine like any (array ['mariadb', '%mysql']) and enabled_cloudwatch_logs_exports ?& array ['audit','error','general','slowquery'] then 'ok'
    when engine like any (array['%postgres%']) and enabled_cloudwatch_logs_exports ?& array ['postgresql','upgrade'] then 'ok'
    when engine like 'oracle%' and enabled_cloudwatch_logs_exports ?& array ['alert','audit', 'trace','listener'] then 'ok'
    when engine = 'sqlserver-ex' and enabled_cloudwatch_logs_exports ?& array ['error'] then 'ok'
    when engine like 'sqlserver%' and enabled_cloudwatch_logs_exports ?& array ['error','agent'] then 'ok'
    else 'alarm'
  end as status,
  case
    when engine like any (array ['mariadb', '%mysql']) and enabled_cloudwatch_logs_exports ?& array ['audit','error','general','slowquery']
    then title || ' ' || engine || ' logging enabled.'
    when engine like any (array['%postgres%']) and enabled_cloudwatch_logs_exports ?& array ['postgresql','upgrade']
    then title || ' ' || engine || ' logging enabled.'
    when engine like 'oracle%' and enabled_cloudwatch_logs_exports ?& array ['alert','audit', 'trace','listener']
    then title || ' ' || engine || ' logging enabled.'
    when engine = 'sqlserver-ex' and enabled_cloudwatch_logs_exports ?& array ['error']
    then title || ' ' || engine || ' logging enabled.'
    when engine like 'sqlserver%' and enabled_cloudwatch_logs_exports ?& array ['error','agent']
    then title || ' ' || engine || ' logging enabled.'
    else title || ' logging not enabled.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_rds_db_instance;

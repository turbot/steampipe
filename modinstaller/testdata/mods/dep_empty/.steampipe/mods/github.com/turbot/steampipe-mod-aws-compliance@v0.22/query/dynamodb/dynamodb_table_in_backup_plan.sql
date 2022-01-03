with mapped_with_id as (
  select
    jsonb_agg(elems) as mapped_ids
  from
    aws_backup_selection,
    jsonb_array_elements(resources) as elems
  group by backup_plan_id
),
mapped_with_tags as (
  select
    jsonb_agg(elems ->> 'ConditionKey') as mapped_tags
  from
    aws_backup_selection,
    jsonb_array_elements(list_of_tags) as elems
  group by backup_plan_id
),
backed_up_table as (
  select
    t.name
  from
    aws_dynamodb_table as t
    join mapped_with_id as m on m.mapped_ids ?| array[t.arn]
  union
  select
    t.name
  from
    aws_dynamodb_table as t
    join mapped_with_tags as m on m.mapped_tags ?| array(select jsonb_object_keys(tags))
)
select
  -- Required Columns
  t.arn as resource,
  case
    when b.name is null then 'alarm'
    else 'ok'
  end as status,
  case
    when b.name is null then t.title || ' not in backup plan.'
    else t.title || ' in backup plan.'
  end as reason,
  -- Additional Dimensions
  t.region,
  t.account_id
from
  aws_dynamodb_table as t
  left join backed_up_table as b on t.name = b.name;

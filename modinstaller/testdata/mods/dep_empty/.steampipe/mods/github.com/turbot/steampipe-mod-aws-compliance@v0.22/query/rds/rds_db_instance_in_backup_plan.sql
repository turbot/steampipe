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
backed_up_instance as (
  select
    i.db_instance_identifier
  from
    aws_rds_db_instance as i
    join mapped_with_id as t on t.mapped_ids ?| array[i.arn]
  union
  select
    i.db_instance_identifier
  from
    aws_rds_db_instance as i
    join mapped_with_tags as t on t.mapped_tags ?| array(select jsonb_object_keys(tags))
)
select
  -- Required Columns
  i.arn as resource,
  case
    when b.db_instance_identifier is null then 'alarm'
    else 'ok'
  end as status,
  case
    when b.db_instance_identifier is null then i.title || ' not in backup plan.'
    else i.title || ' in backup plan.'
  end as reason,
  -- Additional Dimensions
  i.region,
  i.account_id
from
  aws_rds_db_instance as i
  left join backed_up_instance as b on i.db_instance_identifier = b.db_instance_identifier;
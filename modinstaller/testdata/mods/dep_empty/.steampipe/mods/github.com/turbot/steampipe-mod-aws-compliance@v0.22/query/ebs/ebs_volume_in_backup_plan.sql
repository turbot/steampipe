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
backed_up_volume as (
  select
    v.volume_id
  from
    aws_ebs_volume as v
    join mapped_with_id as t on t.mapped_ids ?| array[v.arn]
  union
  select
    v.volume_id
  from
    aws_ebs_volume as v
    join mapped_with_tags as t on t.mapped_tags ?| array(select jsonb_object_keys(tags))
)
select
  -- Required Columns
  v.arn as resource,
  case
    when b.volume_id is null then 'alarm'
    else 'ok'
  end as status,
  case
    when b.volume_id is null then v.title || ' not in backup plan.'
    else v.title || ' in backup plan.'
  end as reason,
  -- Additional Dimensions
  v.region,
  v.account_id
from
  aws_ebs_volume as v
  left join backed_up_volume as b on v.volume_id = b.volume_id;
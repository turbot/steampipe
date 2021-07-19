-- Required Columns
select arn as resource,
       case
           when users is null then 'alarm'
           else 'ok'
           end as status,
       case
           when users is null then title || ' not associated with any IAM user.'
           else title || ' associated with IAM user.'
           end as reason,
       -- Additional Dimensions
       account_id
from
    aws_iam_group;
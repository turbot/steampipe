select
  -- Required Columns
  arn as resource,
  case
    when sse_description is not null and sse_description ->> 'SSEType' = 'KMS' then 'ok'
    when sse_description is null then 'ok'
    else 'alarm'
  end as status,
  case
    when sse_description is not null and sse_description ->> 'SSEType' = 'KMS'
      then title || ' encrypted with AWS KMS.'
    when sse_description is null then title || ' encrypted with DynamoDB managed CMK.'
    else title || ' not encrypted with CMK.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_dynamodb_table;

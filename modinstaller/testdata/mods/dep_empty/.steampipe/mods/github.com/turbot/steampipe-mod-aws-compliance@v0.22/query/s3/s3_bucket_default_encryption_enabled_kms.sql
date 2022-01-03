with data as (
  select
    distinct name
  from
    aws_s3_bucket,
    jsonb_array_elements(server_side_encryption_configuration -> 'Rules') as rules
  where
    rules -> 'ApplyServerSideEncryptionByDefault' ->> 'KMSMasterKeyID' is not null
  )
select
  -- Required Columns
  b.arn as resource,
  case
    when d.name is not null then 'ok'
    else 'alarm'
  end status,
  case
    when d.name is not null then b.name || ' default encryption with KMS enabled.'
    else b.name || ' default encryption with KMS disabled.'
  end reason,
  -- Additional Dimensions
  b.region,
  b.account_id
from
  aws_s3_bucket as b
  left join data as d on b.name = d.name;
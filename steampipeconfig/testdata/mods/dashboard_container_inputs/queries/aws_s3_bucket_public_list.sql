select
    title,
    region,
    account_id
from
    aws_s3_bucket
where
    bucket_policy_is_public
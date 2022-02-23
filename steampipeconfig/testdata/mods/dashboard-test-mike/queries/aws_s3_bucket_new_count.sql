select
    count(*) as "New Buckets"
from
    aws_s3_bucket
where
    creation_date > current_date - 14
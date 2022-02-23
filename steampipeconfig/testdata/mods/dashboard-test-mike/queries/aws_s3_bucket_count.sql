select
    count(*) as "S3 Buckets"
from
    aws_s3_bucket
--where region = 'us-east-1';
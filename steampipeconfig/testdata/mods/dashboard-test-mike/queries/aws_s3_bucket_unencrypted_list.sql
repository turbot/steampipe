select
    title,
    region,
    account_id,
    ('{"link_url":"' || concat('/reports_poc.report.aws_s3_bucket_detail?bucket=',arn) || '"}') :: jsonb as "_ctx[title]"
from
    aws_s3_bucket
where
    server_side_encryption_configuration is null
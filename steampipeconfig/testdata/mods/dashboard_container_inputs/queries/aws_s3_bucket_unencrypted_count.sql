select
    'Unencrypted Buckets' as label,
    count(*) as value,
    case
        when count(*) > 0 then 'alert'
        else 'ok'
    end "type",
    'View Unencrypted Buckets' as link_text,
    '/reports_poc.report.aws_s3_bucket_encryption_report' as link_url
from
    aws_s3_bucket
where
    server_side_encryption_configuration is null
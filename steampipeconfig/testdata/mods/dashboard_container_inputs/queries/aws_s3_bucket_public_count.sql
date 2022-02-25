select
    'Public Buckets' as label,
    count(*) as value,
    case
        when count(*) > 0 then 'alert'
        else 'ok'
    end "type",
    'View Public Buckets' as link_text,
    '/reports_poc.report.aws_s3_bucket_public_access_report' as link_url
from
    aws_s3_bucket
where
    bucket_policy_is_public
select
    case
        when count(*) > 0 then 'alarm'
        else 'ok'
    end status
from
    aws_s3_bucket
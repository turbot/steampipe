select
    region as "Region",
    count(*) as "Total"
from
    aws_s3_bucket
group by
    "Region"
order by
    "Total" desc
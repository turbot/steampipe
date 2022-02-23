select
    region as Region,
    count(*) as Total
from
    aws_kms_key
group by
    region
order by
    Total desc
select
    region as Region,
    count(*) as Total
from
    aws_ec2_instance
group by
    region
order by
    Total desc
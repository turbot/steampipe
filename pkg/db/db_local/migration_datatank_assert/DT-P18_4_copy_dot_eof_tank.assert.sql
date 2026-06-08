-- DT-P18.4 assertion: all 3 rows present, the "\." row's text preserved exactly.
select id, title
from "fast_aws"."aws_resource"
order by id;

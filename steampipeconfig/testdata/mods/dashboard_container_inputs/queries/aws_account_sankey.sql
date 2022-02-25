with analysis as (
    select
        null as parent,
        account_id as id,
        title as name,
        0 as depth,
        'aws_account' as category
    from
        aws_account
    union
    select
        account_id as parent,
        name as id,
        title as name,
        1 as depth,
        'aws_region' as category
    from
        aws_region
    union
    select
        region as parent,
        vpc_id as id,
        title as name,
        2 as depth,
        'aws_vpc' as category
    from
        aws_vpc
    union
    select
        region as parent,
        internet_gateway_id as id,
        title as name,
        2 as depth,
        'aws_vpc_internet_gateway' as category
    from
        aws_vpc_internet_gateway
--  union
--
--  select
--    vpc_id as parent,
--    group_id as id,
--    title as name,
--    3 as depth,
--    'SG' as category
--  from
--    aws_vpc_security_group
--  union
--
--  select
--    group_id as parent,
--    title as id,
--    title as name,
--    4 as depth,
--    'SG Rule' as category
--  from
--    aws_vpc_security_group_rule
    union
    select
        region as parent,
        arn as id,
        title as name,
        2 as depth,
        'aws_vpc_eip' as category
    from
        aws_vpc_eip
    union
    select
        vpc_id as parent,
        subnet_id as id,
        title as name,
        3 as depth,
        'aws_vpc_subnet' as category
    from
        aws_vpc_subnet
    union
    select
        subnet_id as parent,
        arn as id,
        title as name,
        4 as depth,
        'aws_ec2_instance' as category
    from
        aws_ec2_instance
    union
    select
        vpc_id as parent,
        arn as id,
        title as name,
        3 as depth,
        'aws_lambda_function' as category
    from
        aws_lambda_function
    where
        vpc_id is not null
      and vpc_id <> ''
)
select * from analysis order by depth, category, name
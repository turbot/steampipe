with analysis as (
    select
        null as parent,
        account.arn as id,
        account.title as name,
        0 as depth,
        'aws_account' as category
    from
        aws_account account
            inner join
        aws_region region
        on account.account_id = region.account_id
    where
            region.name = 'us-east-1'
    union
    select
        account.arn as parent,
        region.name as id,
        region.title as name,
        1 as depth,
        'aws_region' as category
    from
        aws_region region
    inner join
        aws_account account
    on region.account_id = account.account_id
    where
        name = 'us-east-1'
    union
    select
        region as parent,
        vpc_id as id,
        title as name,
        2 as depth,
        'aws_vpc' as category
    from
        aws_vpc
    where
            region = 'us-east-1'
    union
    select
        sg.vpc_id as parent,
        sg.group_id as id,
        sg.title as name,
        3 as depth,
        'aws_vpc_security_group' as category
    from
        aws_vpc_security_group sg
            inner join
        aws_vpc vpc
        on sg.vpc_id = vpc.vpc_id
        and vpc.region = 'us-east-1'
)
select * from analysis order by depth, category, id;
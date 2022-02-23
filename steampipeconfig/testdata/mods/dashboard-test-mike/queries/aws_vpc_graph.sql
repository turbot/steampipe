with analysis as (
    select
        null as parent,
        region.name as id,
        region.title as name,
        0 as depth,
        'aws_region' as category
    from
        aws_region region
            inner join
        aws_vpc vpc
        on region.region = vpc.region
    where
            vpc.vpc_id = 'vpc-0bf2ca1f6a9319eea'
    union
    select
        region as parent,
        vpc_id as id,
        title as name,
        1 as depth,
        'aws_vpc' as category
    from
        aws_vpc
    where
            vpc_id = 'vpc-0bf2ca1f6a9319eea'
    union
    select
        sg.vpc_id as parent,
        sg.group_id as id,
        sg.title as name,
        2 as depth,
        'aws_vpc_security_group' as category
    from
        aws_vpc_security_group sg
            inner join
        aws_vpc vpc
        on sg.vpc_id = vpc.vpc_id
    where
            sg.vpc_id = 'vpc-0bf2ca1f6a9319eea'
)
select * from analysis order by depth, category, id;
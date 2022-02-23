with analysis as (
    select
        null as parent,
        vpc.vpc_id as id,
        vpc.title as name,
        0 as depth,
        'aws_vpc' as category
    from
        aws_vpc vpc
            inner join
        aws_vpc_security_group sg
        on sg.vpc_id = vpc.vpc_id
    where
            sg.group_name = 'smyth-test-sg1'
    union
    select
        vpc_id as parent,
        group_id as id,
        group_name as name,
        1 as depth,
        'aws_vpc_security_group' as category
    from
        aws_vpc_security_group
    where
            group_name = 'smyth-test-sg1'
--     union
--         select
--          as parent,
--         instance_id as id,
--         2 as depth,
--         'EC2 Instance' as category
--     from
--         aws_ec2_instance
--     where
--         security_groups @>  'smyth-test-sg1'
)
select * from analysis order by depth, category, id;
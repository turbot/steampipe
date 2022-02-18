//
//query "aws_s3_buckets_by_versioning_enabled" {
//    sql = <<-EOQ
//    with versioning as (
//      select
//        case when versioning_enabled then 'Enabled' else 'Disabled' end as versioning_status,
//        region
//      from
//        aws_s3_bucket
//    )
//    select
//      versioning_status,
//      count(versioning_status) as "Total"
//    from
//      versioning
//    where
//      region = $1
//    group by
//      versioning_status
//  EOQ
//    param "region" {
//        default = "us-east-1"
//    }
//}
//
//dashboard "inputs" {
//    title = "Inputs Test"
//
//    chart {
//        type  = "donut"
//        width = 3
//        query = query.aws_s3_buckets_by_versioning_enabled
//        title = "AWS IAM Users MFA Status"
//
//        series "Total" {
//            point "Disabled" {
//                color = "red"
//            }
//
//            point "Enabled" {
//                color = "green"
//            }
//        }
//    }
//}


query "q1"{
    sql = "select {1}"
    param "p1"{
        default = "1"
    }
}

dashboard "r1"{
    input "i1"{
        query = query.q1
        args = {
            "p1" = "FOO"
        }
    }

    chart {
        query = query.q1
        args = {
            "p1" = self.input.i1.value
            "p2" = "foo"
            "p3" = self.input.i1.value
        }
    }

//    control {
//        query = query.q1
//        args = [ self.input.i1.value, "foo", self.input.i1.value]
//    }

}

//dashboard "r2"{
//    dashboard "derived1" {
//        base = dashboard.r1
//    }
//    dashboard "derived2" {
//        base = dashboard.r1
//    }
//}




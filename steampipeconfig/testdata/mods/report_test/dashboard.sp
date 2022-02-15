input "i1"{ }

query "q1"{
    sql = "select 1"
    param "p1"{
    }
}

dashboard "r1"{
    input "i1"{ }

    chart {
        query = query.q1
        args = {
            "p1" = self.input.i1.value
        }
    }

}

//dashboard "r2"{
//    dashboard "derived1" {
//        base = dashboard.r1
//    }
//    dashboard "derived2" {
//        base = dashboard.r1
//    }
//}

query "q1"{
    sql = "select {1}"
    param "p1"{
        default = "1"
    }
}

dashboard "r1"{
    input "i1"{
        query = query.ql
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




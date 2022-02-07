input "i1"{ }

query "q1"{
    sql = "select 1"
    param "p1"{
    }
}

report "r1"{
    input "i1"{ }

    chart {
        query = query.q1
        args = {
            "p1" = self.input.i1.value
        }
    }

}

report "r2"{
    report "derived1" {
        base = report.r1
    }
    report "derived2" {
        base = report.r1
    }
}
input "i1"{ }

report "r1"{
    input "i1"{ }

    chart {
        query = query.q1
        args = {
            "p1" = self.input.i1.value
        }
    }

}

query "q1"{
    sql = "select 1"
    param "p1"{
    }
}
report "r2"{
        base = report.r1
 }
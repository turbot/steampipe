input "i1"{ }

report "r1"{
    input "i1"{ }

    chart {
        query = query.q1
        args = {
            "p1" = "FOO"
        }
    }
     chart {
        query = query.q1
        args = {
            "p1" = self.input.i1.value
        }
    }
     chart {
        query = query.q1
        args = {
            "p1" = input.i1.value
        }
    }
    chart {
           query = query.q1
    }
    chart {
           query = query.q2
    }

}

query "q1"{
    sql = "select 1"
    param "p1"{
    }
}

query "q2"{
    sql = "select 1"
    param "p1"{
        default = "bar"
    }
}

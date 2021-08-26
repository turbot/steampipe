control "c1"{
    title ="C1"
    description = "THIS IS CONTROL 1"
    sql = "query.q1"
}

control "c2"{
    title ="C2"
    description = "THIS IS CONTROL 2"
    sql = "query.q1"
    params = {
        "p1" = "control2 "
        "p3" = "a reason"
    }
}

control "c3"{
    title ="C3"
    description = "THIS IS CONTROL 3"
    sql = "query.q1"
    params = [  "control3 ", "because______ " ]
}

control "c1"{
    title ="C1"
    description = "THIS IS CONTROL 1"
    query = query.q1
}

control "c2"{
    title ="C2"
    description = "THIS IS CONTROL 2"
    query = query.q1
    params = {
        "p1" = "control2 "
        "p2" = 100
        "p3" = "a reason"
    }
}

control "c3"{
    title ="C3"
    description = "THIS IS CONTROL 3"
    query = query.q1
    params = [  "control3 ", "because______ " ]
}

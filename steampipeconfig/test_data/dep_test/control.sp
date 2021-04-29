
control "c1"{
    sql = "select 1"
    labels = ["public cloud", "aws"]
    title = upper("aws cis")
}
control "c2"{
    sql = "select 2"
    labels = ["public cloud", "aws"]
    title = upper("aws cis")
}
control "c3"{
    sql = "select 3"
    labels = ["public cloud", "aws"]
    title = upper("aws cis")
}
control "c4"{
    sql = control.c1.sql
    labels = ["public cloud", "aws"]
    title = upper("aws cis")
}
control "c5"{
    sql = control.c2.sql
    labels = ["public cloud", "aws"]
    title = upper("aws cis")
}
control "c6"{
    sql = control.c3.sql
    labels = ["public cloud", "aws"]
    title = upper("aws cis")
}

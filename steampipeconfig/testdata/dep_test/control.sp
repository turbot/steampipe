
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
    sql = control.c1.name
    labels = concat(control.c3.labels,[
"cis_item_id:1"
])
    title = control.c1.name
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

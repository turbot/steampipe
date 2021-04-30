
control "c1"{
    sql = control.c2.sql
    labels = ["public cloud", "aws"]
    title = control.c2.name
}
control "c2"{
    sql = "SELECT 1"
    labels = ["public cloud", "aws"]
    title = "TITLE"
}
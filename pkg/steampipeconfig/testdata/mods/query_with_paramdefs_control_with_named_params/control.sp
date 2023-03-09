control "c1"{
    title ="C1"
    description = "THIS IS CONTROL 1"
    sql = "select 'ok' as status, $1 as resource, $2 as reason"
    param "p1" {
        default = "val1"
    }
    param "p2" {
        default = "val2"
    }
    args = [["my val1","boo"], "my val2"]
}

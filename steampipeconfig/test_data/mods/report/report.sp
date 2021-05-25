report "r1"{}

panel "p1"{
    title = "foo"
    width = 100
    source = "THIS IS A PANEL OK"
    sql = "select 1"
    panel "p2"{
        title = "bar"
        width = 200
        source = "THIS IS A PANEL OK"
        sql = "select 1"
    }
}

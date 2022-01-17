report "r1" {
    container {
        panel  {
            title = "foo"
            width = 100
            height = 10
            source = "THIS IS A PANEL OK"
            sql = "select 1"
        }
    }
}

panel "name" {
    title = "foo"
    width = 100
    height = 10
    source = "THIS IS A PANEL OK"
    sql = "select 1"
}


container "foo"{
    panel  {
        title = "foo"
        width = 100
        height = 10
        source = "THIS IS A PANEL OK"
        sql = "select 1"
    }
}
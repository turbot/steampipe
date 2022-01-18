report "r1" {

        base = container.foo
}

panel "name" {
    title = "foo"
    width = 100
    height = 10
    source = "THIS IS A PANEL OK"
    sql = "select 1"
}



container "foo" {
    panel {
        type = "markdown"
        text = "## Some title1"
    }
}
report "r1" {
    container {
        panel  {
            title = "foo"
            width = 100
            height = 10
            type = "markdown"
            text = "## Some title a"
        }
        panel {
            type = "markdown"
            text = "## Some title b"
        }
        panel {
            type = "markdown"
            text = "## Some title c"
        }
        panel {
            type = "markdown"
            text = "## Some title d"
        }
        container{
            base = container.foo
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
    container {
        container {
            container {
                panel {
                    type = "markdown"
                    text = "## Some title1"
                }
                panel {
                    type = "markdown"
                    text = "## Some title2"
                }
            }
        }
    }
}
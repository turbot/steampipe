report "r1" {
    panel "p1" {
        title = "foo"
        width = 100
        height = 10
        source = "THIS IS A PANEL OK"
        sql = "select 1"
        panel "p1_1" {
            title = "bar"
            width = 200
            height = 20
            source = "THIS IS A PANEL OK"
            sql = "select 1"
        }
        panel "p1_2" {
            title = "bar"
            width = 200
            height = 20
            source = "THIS IS A PANEL OK"
            report "nested" {
                panel "p1_2_1" {
                    title = "boobar"
                    width = 200
                    source = "THIS OK"
                    sql = "select 1"
                }
            }
        }
    }
}
report "r1" {

        base = container.foo
}
//
//counter "name" {
//    title = "foo"
//    width = 100
//    sql = "select 1"
//}



container "foo" {
    text {
        value = "SOME TEXT"
    }
}
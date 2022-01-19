report "r1" {

        base = container.foo
}
report "r2" {


  container  {
    text {
      type = "markdown"
      value = "SOME OTHER TEXT"
    }
  }
}
//
//counter "name" {
//    title = "foo"
//    width = 100
//    sql = "select 1"
//}




container "foo" {
    text {
        type = "markdown"
        value = "SOME TEXT"
    }
}
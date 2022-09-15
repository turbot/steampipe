// this dashboard is used to test the parsing of a simple dashboard

dashboard "simple_report" {
  text {
    value = "a simple report"
  }

  chart {
    title = "a simple query"
    query = query.q1
  }


}


query "q1"{
    title ="Q1"
    description = "THIS IS QUERY 1"
    sql = "selt 1 "
     param "p1"{
            description = "desc"
            default = "I am dt"
        }
}

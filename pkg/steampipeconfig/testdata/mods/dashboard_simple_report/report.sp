// this dashboard is used to test the parsing of a simple dashboard

dashboard "simple_report" {
  text {
    value = "a simple report"
  }

  chart {
    title = "a simple query"
    sql = "select 1"
  }
}

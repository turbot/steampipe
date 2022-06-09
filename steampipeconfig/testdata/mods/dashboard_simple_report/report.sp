// this dashboard is used to test the parsing of a simple dashboard

dashboard "simple_report" {
  text {
    value = "a simple report"
  }

  container {
      chart "c1"{
        title = "a simple query"
        sql = "select 1"
      }
      chart "c1"{
        title = "a simple query"
        sql = "select 1"
      }
  }
}
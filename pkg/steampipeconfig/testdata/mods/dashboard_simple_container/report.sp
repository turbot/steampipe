// this dashboard is used to test the parsing of a simple container

dashboard "simple_container_report" {
  container {
    text {
      value = "container 1"
    }
    chart {
      title = "container 1 chart 1"
      sql = "select 1 as container"
    }
  }
}
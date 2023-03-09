// this dashboard is used to test the parsing of a dashboard containing
// multiple sibling containers

dashboard "sibling_containers_report" {
  container {
    text {
      value = "container 1"
    }
    chart {
      title = "container 1 chart 1"
      sql = "select 1 as container"
    }
  }

  container {
    text {
      value = "container 2"
    }
    chart {
      title = "container 2 chart 1"
      sql = "select 2 as container"
    }
  }

  container {
    text {
      value = "container 3"
    }
    chart {
      title = "container 3 chart 1"
      sql = "select 3 as container"
    }
  }
}
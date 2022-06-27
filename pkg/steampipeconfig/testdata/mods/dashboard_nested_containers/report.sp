// this dashboard is used to test the parsing of a dashboard containing
// nested containers

dashboard "nested_containers_report" {
  container {
    text {
      value = "CONTAINER 1"
    }
    container {
      text {
        value = "CHILD CONTAINER 1.1"
      }
      chart {
        title = "CHART 1"
        sql = "select 1.1 as container"
      }
    }
    container {
      text {
        value = "CHILD CONTAINER 1.2"
      }
      chart {
        title = "CHART 2"
        sql = "select 1.2 as container"
      }
      container {
        text {
          value = "NESTED CHILD CONTAINER 1.2.1"
        }
        chart {
          title = "CHART 3"
          sql = "select 1.2.1 as container"
        }
      }
    }
  }
}
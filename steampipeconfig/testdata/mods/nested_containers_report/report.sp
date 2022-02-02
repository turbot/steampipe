report "nested_containers_report" {
  container {
    text {
      value = "CONTAINER 1"
    }
    container {
      text {
        value = "CHILD CONTAINER 1(1)"
      }
      chart {
        title = "CHART 1"
        sql = "select 1 as child_container, 1 as container"
      }
    }
    container {
      text {
        value = "CHILD CONTAINER 2(1)"
      }
      chart {
        title = "CHART 2"
        sql = "select 2 as child_container, 1 as container"
      }
      container {
        text {
          value = "NESTED CHILD CONTAINER 1(21)"
        }
        chart {
          title = "CHART 3"
          sql = "select 1 as child_container, 2 as container"
        }
      }
    }
  }
}
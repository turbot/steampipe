// this dashboard is a simple dashboard containing charts with axes.
// we are testing the parsing and the inheritance of the base values.

query basic_query {
  sql = "select 1"
}

chart basic_chart {
  type = "column"
  sql = query.basic_query.sql
  grouping = "compare"
  legend {
    position = "bottom"
  }
  axes {
    x {
      title {
        display = "always"
        value = "Foo"
      }
    }
    y {
      title {
        display = "always"
        value = "Foo"
      }
    }
  }
}

dashboard inheriting_from_base {
  title = "inheriting_from_base"

  chart {
    base = chart.basic_chart
    width = 8
    axes {
      x {
        title {
          value = "Barz"
        }
      }
    }
  }
}
chart "top_level1" {
  title = "top level 1"
  sql = "select 1 as chart"
}

chart "top_level2" {
  title = "top level 2"
  sql = "select 2 as chart"
}

dashboard "anonymous_naming" {

  chart {
    title = "chart within dashboard"
    sql = "select 3 as chart"
  }

  container {
    chart {
      title = "chart 1.1"
      sql = "select 4 as chart"
    }
    chart {
      title = "chart 1.2"
      sql = "select 5 as chart"
    }
    table {
      title = "table 1.1"
      sql = "select 1 as table"
    }
  }

  container {
    chart {
      title = "chart 2.1"
      sql = "select 6 as chart"
    }
    chart {
      title = "chart 2.2"
      sql = "select 7 as chart"
    }
    table {
      title = "table 2.1"
      sql = "select 2 as table"
    }
    table {
      title = "table 2.2"
      sql = "select 3 as table"
    }
  }
}
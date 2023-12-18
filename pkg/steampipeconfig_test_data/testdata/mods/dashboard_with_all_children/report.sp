// this dashboard contains all possible child resources
// we are testing the parsing of all possible child resources
// TODO add input block in dashboard

dashboard dashboard_with_child_res {
  title = "dashboard with child resources"

  container {
    title = "example container"
  }
  chart {
    title = "example chart"
    sql = "select 1"
  }
  card {
    title = "example card"
    sql = "select 1"
    type = "ok"
  }
  flow {
    title = "example flow"
    type = "sankey"
  }
  graph {
    title = "example graph"
    type = "graph"
  }
  hierarchy {
    title = "example hierarchy"
    type = "graph"
  }
  image {
    title = "example image"
    src = "https://steampipe.io/images/logo.png"
    alt = "steampipe"
  }
  input "i1" {
    title = "example input"
  }
  table {
    title = "example table"
    sql = "select 1"
  }
  text {
    value = "example text"
  }
}
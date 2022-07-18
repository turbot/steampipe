// this dashboard contains a simple container with all possible child resources
// we are testing the parsing of all possible child resources
// TODO add input block in container

dashboard container_with_child_res {
  title = "container with child resources"

  container {
    title = "example container with all possible child resources"

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
    table {
      title = "example table"
      sql = "select 1"
    }
    text {
      value = "example text"
    }
  }
}
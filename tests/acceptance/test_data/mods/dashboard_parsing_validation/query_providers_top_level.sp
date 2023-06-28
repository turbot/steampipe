dashboard "query_providers_top_level" {
  title = "Query providers at top level that do not require a query/sql block"
  description = "This is a dashboard that validates - Query providers at top level DO NOT need a query/sql block"

  card "card_1" {
    base = card.top_card
  }

  flow "flow1" {
    base = flow.top_flow
  }

  graph "graph_1" {
    base = graph.top_graph
  }

  hierarchy "hierarchy_1" {
    base = hierarchy.top_hierarchy
  }

  image "image_1" {
    base = image.top_image
  }

  input "input_1" {
    base = input.top_input
  }
}

card "top_card" {
  width = 2
  label = "Card"
  value = "TopLevelCard"
}

chart "chart_top_1" {
  width = 5
  title = "Top level Chart"
}

flow "top_flow" {
  title = "TopLevelFlow"
  width = 5

  node "node_flow_1" {
    sql = <<-EOQ
      select 1 as node
    EOQ
  }
  edge "edge_flow_1" {
    sql = <<-EOQ
      select 1 as edge
    EOQ
  }
}

graph "top_graph" {
  title = "Top level graph"
  width = 5

  node "node_graph_1" {
    sql = <<-EOQ
      select 1 as node
    EOQ
  }
  edge "edge_graph_1" {
    sql = <<-EOQ
      select 1 as edge
    EOQ
  }
}

hierarchy "top_hierarchy" {
  title = "Top level hierarchy"
  width = 5

  node "node_hierarchy_1" {
    sql = <<-EOQ
      select 1 as node
    EOQ
  }
  edge "edge_hierarchy_1" {
    sql = <<-EOQ
      select 1 as edge
    EOQ
  }
}

image "top_image" {
  title = "top level image"
  width = 3
  src = "https://steampipe.io/images/logo.png"
  alt = "steampipe"
}

input "top_input" {
  width = 2
  type = "text"
  display = "TopLevelInput"
}

table "top_table" {
  width = 4
  display = "TopLevelTable"
}
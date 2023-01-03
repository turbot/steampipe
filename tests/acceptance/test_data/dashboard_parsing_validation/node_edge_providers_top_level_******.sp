dashboard "node_edge_providers_top_level" {
  title = "Node and Edge providers at top level do not need query/sql block or node/edge blocks"
  description = "This is a dashboard that validates - Node and Edge providers at top level DO NOT need query/sql block or node/edge blocks"

  flow "flow1" {
    base = flow.top_flow_1
  }

  graph "graph_1" {
    base = graph.top_graph_1
  }

  hierarchy "hierarchy_1" {
    base = hierarchy.top_hierarchy_1
  }
}

flow "top_flow_1" {
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

graph "top_graph_1" {
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

hierarchy "top_hierarchy_1" {
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

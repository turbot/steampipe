dashboard "node_edge_providers_nested" {
  title = "Node and Edge providers(nested) that always require a query/sql block or a node/edge"
  description = "This is a dashboard that validates - nested Node and Edge providers always need a query/sql block or a node/edge block"

  container {
    flow "nested_flow_1" {
      title = "Nested flow"
      width = 3

      node "node_nested_flow" {
        sql = <<-EOQ
          select 1 as node
        EOQ
      }
      edge "edge_nested_flow" {
        sql = <<-EOQ
          select 1 as edge
        EOQ
      }
    }

    graph "nested_graph_1" {
      title = "Nested graph"
      width = 5

      node "node_nested_graph" {
        sql = <<-EOQ
          select 1 as node
        EOQ
      }
      edge "edge_nested_graph" {
        sql = <<-EOQ
          select 1 as edge
        EOQ
      }
    }

    hierarchy "nested_hierarchy_1" {
      title = "Nested hierarchy"
      width = 5

      node "node_nested_hierarchy" {
        sql = <<-EOQ
          select 1 as node
        EOQ
      }
      edge "edge_nested_hierarchy" {
        sql = <<-EOQ
          select 1 as edge
        EOQ
      }
    }
  }
}
dashboard "node_edge_providers_nested" {
  title = "Node and Edge providers(nested) always require a query/sql block or a node/edge block"
  description = "This is a dashboard that validates - nested Node and Edge providers always need a query/sql block or a node/edge block - SHOULD RESULT IN PARSING FAILURE"

  container {
    flow "nested_flow_1" {
      title = "Nested flow"
      width = 3
    }

    graph "nested_graph_1" {
      title = "Nested graph"
      width = 5
    }

    hierarchy "nested_hierarchy_1" {
      title = "Nested hierarchy"
      width = 5
    }
  }
}
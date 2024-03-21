dashboard "nested_dashboards" {
  title = "Nested dashboards"
  dashboard "reused_node_edge_providers_nested" {
    base = dashboard.node_edge_providers_nested
  }
}
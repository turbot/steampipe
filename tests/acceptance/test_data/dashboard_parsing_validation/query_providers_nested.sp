dashboard "query_providers_nested" {
  title = "Query providers(nested) that always require a query/sql block"
  description = "This is a dashboard that validates - nested Query providers always need a query/sql block"

  container {
    chart "nested_chart" {
      sql = "select 1 as chart"
      width = 5
      title = "Nested Chart"
    }

    flow "nested_flow" {
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

    graph "nested_graph" {
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

    hierarchy "nested_hierarchy" {
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

    table "nested_table" {
      sql = "select 1 as table"
      width = 4
      title = "Nested table"
    }

    # input type="text" does not require a query/sql block,
    # anything other than that requires a query/sql
    input "nested_input" {
      sql = "select 1 as input"
      width = 2
      title = "Nested input"
    }

    input "nested_input_type_text" {
      type = "text"
      width = 2
      title = "Nested input type text"
    }
  }
}
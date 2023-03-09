dashboard "query_providers_nested" {
  title = "Query providers(nested) always require a query/sql block"
  description = "This is a dashboard that validates - nested Query providers always need a query/sql block - SHOULD RESULT IN PARSING FAILURE"

  container {
    chart "nested_chart" {
      width = 5
      title = "Nested Chart"
    }

    table "nested_table" {
      width = 4
      title = "Nested table"
    }
  }
}
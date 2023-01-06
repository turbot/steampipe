dashboard "query_providers_nested_dont_require_sql" {
  title = "Query providers(nested) that do not require a query/sql block"
  description = "This is a dashboard that validates - nested Query providers like image and card do not need a query/sql block"

  container {
    image "nested_image" {
      title = "Nested image"
      width = 3
      src = "https://steampipe.io/images/logo.png"
      alt = "steampipe"
    }

    card "nested_card" {
      width = 2
      label = "Card"
      value = "Nested Card"
    }
  }
}
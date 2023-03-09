

dashboard "testing_dashboard_inputs" {

  title         = "Dashboard input testing"

  input "new_input" {
    title       = "Enter a text:"
    width       = 4
    type        = "text"
  }

  table {
    type  = "line"
    query = query.query_input
    args  = {
      new_input = self.input.new_input.value
    }

    column "Alternative Names" {
      wrap = "all"
    }
  }
}

query "query_input" {
  sql = <<-EOQ
    select
      'value1' as "column 1",
      'value1' as "column 2"
  EOQ
}
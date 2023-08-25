
input "base_input" {
  title = "Select resource compliance state"
  width = 4
  type  = "select"

  option "compliant" {
    label = "Compliant"
  }

  option "non-compliant" {
    label = "Non-Compliant"
  }
}


dashboard "resource_details" {
  title = "Resource Details"

  input "resource_compliance_state" {
    base = input.base_input
  }

  table {
    width = 12
    sql   = "select 1"
  }
}
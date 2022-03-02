input "global_input" {
  title = "example global input"
}

dashboard "global_dashboard_inputs" {
  title = "global dashboard inputs"

  input "i1" {
    base = input.global_input
  }
}
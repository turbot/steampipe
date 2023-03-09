dashboard dashboard_with_duplicate_inputs {
  title = "dashboard with duplicate inputs"

  input "i1" {
    title = "example input 1"
  }
  input "i1" {
    title = "example input 2"
  }
}
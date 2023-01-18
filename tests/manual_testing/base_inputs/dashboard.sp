dashboard "base_inputs" {
  input "input_1" {
    base = input.top_input
  }
}


input "top_input" {
  width = 2
  type = "text"
  display = "TopLevelInput"
}

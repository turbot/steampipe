
input "global_input"{
  title = "global input"
}

dashboard "d1" {
  title = input.global_input
}
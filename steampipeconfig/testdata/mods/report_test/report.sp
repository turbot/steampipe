report "r1" {
  text {
      value = "hi you"
  }
  container "container1" {
    base  = container.some_other_container
    width = 6
  }
}
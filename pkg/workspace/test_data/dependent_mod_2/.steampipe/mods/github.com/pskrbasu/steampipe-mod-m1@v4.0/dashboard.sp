chart global_chart {
    title = "chart global m1_c1"
    sql = "select 'm1_global_c1' as chart"
}

input "global_input" {
    title = "global input m1_i1"
}


dashboard m1_d1 {
  title = "dashboard d1"

  chart {
    title = "chart c1"
    sql = "select 'm1_d1_c1' as chart"
  }
  
  input "i1" {
    title = "input i1"
  }
}

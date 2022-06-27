chart chart1{
  title = "chart 1"
  query = m1.query.m1_q1
}

dashboard local_d1 {
  title = "dashboard d1"

  chart {
    base = m1.chart.global_chart
  }
  
  input "i1" {
    base = m1.input.global_input
  }
}
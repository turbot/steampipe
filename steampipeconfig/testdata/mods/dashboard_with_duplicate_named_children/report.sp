// dashboard with duplicate named child block

dashboard dashboard_with_child_res {
  title = "dashboard with child resources"

  container "cnt1" {
    title = "example container"
  }
  chart "c1" {
    title = "example chart"
    sql = "select 1"
  }
  chart "c1" {
    title = "example chart"
    sql = "select 1"
  }
}

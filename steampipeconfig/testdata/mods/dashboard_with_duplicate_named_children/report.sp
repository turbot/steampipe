// this dashboard contains all possible child resources
// we are testing the parsing of all possible child resources
// TODO add input block in dashboard

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
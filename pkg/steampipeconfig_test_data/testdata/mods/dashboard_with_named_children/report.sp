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
  card "crd1"{
    title = "example card"
    sql = "select 1"
    type = "ok"
  }
  flow "f1"{
    title = "example flow"
    type = "sankey"
  }
  graph "g1"{
    title = "example graph"
    type = "graph"
  }
  hierarchy "h1" {
    title = "example hierarchy"
    type = "graph"
  }
  image "i1"{
    title = "example image"
    src = "https://steampipe.io/images/logo.png"
    alt = "steampipe"
  }
  input "ip1" {
    title = "example input"
  }
  table "t1"{
    title = "example table"
    sql = "select 1"
  }
  text "txt1"{
    value = "example text"
  }
}
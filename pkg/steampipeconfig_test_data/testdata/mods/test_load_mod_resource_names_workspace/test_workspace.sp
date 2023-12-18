benchmark "test_workspace" {
  title = "Sample benchmark for unit tests"
  children = [
    control.test_workspace_1,
    control.test_workspace_2,
    control.test_workspace_3
  ]
}

control "test_workspace_1" {
  title = "Sample control 1"
  description = "Sampple control 1"
  sql = query.query_control_1.sql
}

control "test_workspace_2" {
  title = "Sample control 2"
  description = "Sampple control 2"
  sql = query.query_control_2.sql
}

control "test_workspace_3" {
  title = "Sample control 3"
  description = "Sampple control 3"
  sql = query.query_control_3.sql
}

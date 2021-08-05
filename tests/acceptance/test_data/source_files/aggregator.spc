connection "chaos" {
    plugin = "chaos"
}

connection "chaos2" {
  plugin = "chaos"
}

connection "chaos_group" {
  type        = "aggregator"
  plugin      = "chaos"
  connections = ["*"]
}
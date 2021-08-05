connection "chaos_group" {
  type        = "aggregator"
  plugin      = "chaos"
  connections = ["*"]
}
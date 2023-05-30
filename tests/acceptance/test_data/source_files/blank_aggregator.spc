connection "all_chaos" {
  type        = "aggregator"
  plugin      = "chaos"
  connections = ["*"]
}

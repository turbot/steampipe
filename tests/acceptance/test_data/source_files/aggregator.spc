connection "chaos_01" {
  plugin      = "chaos" 
  profile     = "chaos_01"
}

connection "chaos_02" {
  plugin      = "chaos" 
  profile     = "chaos_02"
}

connection "steampipe_01" {
  plugin      = "steampipe" 
  profile     = "steampipe_02"
}

connection "chaos_group" {
  type        = "aggregator"
  plugin      = "chaos"
  connections = ["*"]
}
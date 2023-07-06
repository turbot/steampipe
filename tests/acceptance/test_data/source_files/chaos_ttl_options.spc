connection "chaos_ttl_options" {
    plugin = "chaos"
    options "connection" {
      cache = true
      cache_ttl = 10
    }
}
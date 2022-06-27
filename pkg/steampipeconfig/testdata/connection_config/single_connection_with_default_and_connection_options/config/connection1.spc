connection "a" {
  plugin = "test_data/connection-test-1"

  options "connection" {
     cache     = true # true, false
     cache_ttl = 300  # expiration (TTL) in seconds
   }
}

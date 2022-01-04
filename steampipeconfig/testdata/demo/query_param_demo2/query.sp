query "bad_query" {
    sql = <<-EOT
    this is invalid

  EOT
    param "tag_keys" {
        default     = "true"
    }
}
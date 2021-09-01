query "expired_access_keys" {
    sql = <<-EOT

    select
      akas ->> 0 as resource,
      case
          when create_date > NOW() - ($1 || ' days')::interval then 'ok'
          else 'alarm'
      end as status,
      access_key_id || 'for user ' || user_name || ' is ' || age(create_date) || ' old.' as reason,
      region,
      account_id
    from
      aws_iam_access_key

  EOT
    param "max_days" {
        description = "The maximum number of days a key is allowed to exist after it is created."
        default     = 90
    }
}

control "expired_access_keys" {
    title       = "Expired IAM Access Keys"
    query       = query.expired_access_keys
    args        = {
        "max_days"   = 365
    }
}
mod "test_compliance" {
  # hub metadata
  title       = "Test Compliance"
  description = "Test Compliance"
}

variable "allowed_ips" {
  type        = list(number)
  default     = []
  description = "A list of IPs allowed in Snowflake network policies."
}

control "security_overview_network_security_network_policy_allowed_list_set" {
  title       = "Use network policies to allow 'known' client locations (IP ranges)"
  description = "TO DO."
  sql         = <<-EOQ
  with applied_network_policy as (
    select
      'sample' as name,
      array['10.255.255.255', '172.31.255.255', '192.168.255.255'] as allowed_ip_list,
      'test' as account
  ),
  analysis as (
    select
      name,
      to_jsonb ($1::text[]) <@ array_to_json(allowed_ip_list)::jsonb as has_allowed_ips,
      to_jsonb ($1::text[]) - allowed_ip_list as missing_ips,
      account
    from
      applied_network_policy
  )
  select
    -- Required columns
    name as resource,
    case when has_allowed_ips then 'ok' else 'alarm' end as status,
    missing_ips as reason,
    -- Additional columns
    account
  from
    analysis
  EOQ

  param "allowed_ips" {
    default = var.allowed_ips
  }
}



control "c2" {
  title       = "Use network policies to allow 'known' client locations (IP ranges)"
  description = "TO DO."
  sql         = <<-EOQ

  select
    'me' as resource,
    'alarm' as status,
    ($1::text[]) as reason
  EOQ

  param "allowed_ips" {
    default = var.allowed_ips
  }
}

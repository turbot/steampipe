mod "local" {
  title       = "Test Compliance"
  description = "Test Compliance"
}

variable "string_list" {
  type        = list(string)
  default     = []
  description = "A list of strings."
}

control "the_control" {
  title       = "Sample control to test empty list in HCL"
  description = ""
  sql         = <<-EOQ
  with applied_network_policy as (
    select
      'sample' as name,
      array['a', 'dummy', 'list'] as allowed_ip_list,
      'test' as account
  ),
  analysis as (
    select
      name,
      to_jsonb ($1::text[]) <@ array_to_json(allowed_ip_list)::jsonb as has_string_list,
      to_jsonb ($1::text[]) - allowed_ip_list as missing_ips,
      account
    from
      applied_network_policy
  )
  select
    -- Required columns
    name as resource,
    case when has_string_list then 'ok' else 'alarm' end as status,
    missing_ips as reason,
    -- Additional columns
    account
  from
    analysis
  EOQ

  param "string_list" {
    default = var.string_list
  }
}

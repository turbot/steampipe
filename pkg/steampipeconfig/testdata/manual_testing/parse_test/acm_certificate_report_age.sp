
query "aws_acm_certificate_age_table" {
  sql = <<-EOQ
    select
      c.domain_name as "Domain Name",
      c.title as "ID",
      now()::date - c.created_at::date as "Age in Days",
      c.created_at as "Create Time",
      c.not_after as "Expiry Time",
      c.status as "Status",
      a.title as "Account",
      c.account_id as "Account ID",
      c.region as "Region",
      c.certificate_arn as "ARN"
    from
      aws_acm_certificate as c,
      aws_account as a
    where
      c.account_id = a.account_id
    order by
      c.domain_name;
  EOQ
}


dashboard "acm_certificate_age_report" {

  container {

    card {
      query = query.aws_acm_certificate_age_table
      width = 2
    }
  }
}

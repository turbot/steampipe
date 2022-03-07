mod "dashboard_poc" {
  title = "Dashboard POC"
}


dashboard "aws_test" {

  input "account_id" {
    title = "Account"
    width = 2
    query = query.aws_accounts
  }

  input "versioning_enabled" {
    title = "Versioning Enabled?"
    width = 2
    option "true" {}
    option "false" {}
  }


  table {
    title = "named args"
    width = 12
    args = {
      account_id = self.input.account_id.value,
      versioning_enabled = self.input.versioning_enabled.value
    }
    query = query.s3_data
  }

  table {
    title = "inline"
    width = 12
    args = [
      self.input.account_id.value,
      self.input.versioning_enabled.value
    ]
    sql = <<EOT
        select
            title,
            versioning_enabled,
            bucket_policy_is_public
        from
            aws_s3_bucket
        where
          account_id =  $1
          and versioning_enabled  = $2::boolean
      EOT

  }

}

query "s3_data" {
  sql = <<EOT
        select
            title,
            versioning_enabled,
            bucket_policy_is_public
        from
            aws_s3_bucket
        where
          account_id = $1
          and versioning_enabled  = $2::boolean
      EOT
  param "account_id" {}
  param "versioning_enabled" {}
}

query "aws_accounts" {
  sql   = "select title as label, account_id as value from aws_account"
}
query "test" {
  sql = "select _ctx, account_aliases,arn from aws_account"
  search_path_prefix = "aws"

}
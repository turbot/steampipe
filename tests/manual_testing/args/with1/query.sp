query "array_arg" {
  description = "test array argument"
  sql = <<-EOQ
      select
       name
      from
        aws_iam_user
      where
        arn = any($1::text[]);

      EOQ

  param arns{
    default = [
      "arn:aws:iam::876515858155:user/lalit",
      "arn:aws:iam::876515858155:user/mike"
    ]
  }
}

query "single_arg" {
  description = "single arg"
  sql = "select $1"

  param p1{
    default = "foo"
  }
}
query "q1" {
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
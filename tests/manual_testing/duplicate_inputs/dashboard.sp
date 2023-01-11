dashboard "d1" {
  title = "Inputs"

  input "i1" {
    sql = <<-EOQ
          select arn as label, arn as value from aws_account
        EOQ
  }


}

dashboard "d2" {
  title = "Inputs"

  input "i1" {
    sql = <<-EOQ
          select arn as label, arn as value from aws_account
        EOQ
  }


}

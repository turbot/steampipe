card "aws_iam_user_mfa_for_user" {
    sql = query.aws_iam_user_mfa_for_user.sql
    //query = query.aws_iam_user_mfa_for_user

    args = {
        arn = self.input.user_arn.value
    }
    width = 2

}


dashboard "aws_iam_user_detail" {
    title = "AWS IAM User Detail"

    chart {
        sql = query.aws_iam_user_input.sql
        width = 2
        series "count" {
            point "Enabled" {
                color = "green"
            }

            point "Disabled" {
                color = "red"
            }
        }
    }

}


query "aws_iam_user_mfa_for_user" {
    sql = "select 1"


}

query "aws_iam_user_input" {
    sql = "select 1"
}
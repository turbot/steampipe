## Description

This control checks whether Amazon S3 buckets have policies that require requests to use Secure Socket Layer (SSL).

S3 buckets should have policies that require all requests (Action: S3:*) to only accept transmission of data over HTTPS in the S3 resource policy, indicated by the condition key aws:SecureTransport.

This does not check the SSL or TLS version. You should not allow early versions of SSL or TLS (SSLv3, TLS1.0) per PCI DSS requirements.

## Remediation

1. Open the [Amazon S3 console](https://console.aws.amazon.com/s3/).
2. Navigate to the noncompliant bucket, and then choose the bucket name.
3. Choose **Permissions**, then choose **Bucket Policy**.
4. Add a similar policy statement to that in the policy below. Replace `awsexamplebucket` with the name of the bucket you are modifying.

```json
{
    "Id": "ExamplePolicy",
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "AllowSSLRequestsOnly",
            "Action": "s3:*",
            "Effect": "Deny",
            "Resource": [
                "arn:aws:s3:::awsexamplebucket",
                "arn:aws:s3:::awsexamplebucket/*"
            ],
            "Condition": {
                "Bool": {
                     "aws:SecureTransport": "false"
                }
            },
           "Principal": "*"
        }
    ]
}
```

5. Choose **Save**.
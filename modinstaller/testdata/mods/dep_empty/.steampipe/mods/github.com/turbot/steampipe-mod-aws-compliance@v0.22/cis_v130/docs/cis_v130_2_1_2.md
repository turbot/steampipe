## Description

Amazon S3 provides multiple encryption options to protect data at rest, transit & it's access. At the Amazon S3 bucket level, you can restrict bucket policy making the objects accessible only through HTTPS.

By default, Amazon S3 allows both HTTP and HTTPS requests. To achieve only allowing access to Amazon S3 objects through HTTPS you also have to explicitly deny access to HTTP requests. Bucket policies that allow HTTPS requests without explicitly denying HTTP requests will not comply with this recommendation.

## Remediation

### From Console

1. Open the Amazon S3 console [S3](https://console.aws.amazon.com/s3/)
2. Select the **Check box** next to the Bucket.
3. Click on **Permissions**.
4. Click **Bucket Policy**
5. Add this to the existing policy filling in the required information
```
{
   "Sid":"<optional>",
   "Effect":"Deny",
   "Principal":"*",
   "Action":"s3:GetObject",
   "Resource":"arn:aws:s3:::<bucket_name>/*",
   "Condition":{
      "Bool":{
         "aws:SecureTransport":"false"
      }
..
```
6. Choose **Save**
7. Repeat for all the buckets in your AWS account that contain sensitive data.

### Using AWS Policy Generator

1. Repeat steps 1-4 above.
2. Click on **Policy Generator** at the bottom of the Bucket Policy Editor
3. Select Policy Type `S3 Bucket Policy`
4. Add Statements 
    - Effect = Deny 
    - Principal = * 
    - AWS Service = Amazon S3 
    - Actions = GetObject 
    - Amazon Resource Name =
5. Generate Policy
6. Copy the text and add it to the Bucket Policy.

### From Command Line

1. Export the bucket policy to a json file.

```bash
aws s3api get-bucket-policy --bucket <bucket_name> --query Policy --output text > policy.json
```

2. Modify the policy.json file by adding in this statement

```
{
            "Sid": <optional>",
            "Effect": "Deny",
            "Principal": "*",
            "Action": "s3:GetObject",
            "Resource": "arn:aws:s3:::<bucket_name>/*",
            "Condition": {
                "Bool": {
                    "aws:SecureTransport": "false"
                }
            }
        }
```

3. Apply this modified policy back to the S3 bucket:

```bash
aws s3api put-bucket-policy --bucket <bucket_name> --policy file://policy.json
```

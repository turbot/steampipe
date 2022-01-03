## Description

SSL/TLS server certificate is required to enable HTTPS connections to your website or application in AWS. You can use ACM or IAM to store and deploy server certificates. IAM securely encrypts your private keys and stores the encrypted version in IAM SSL certificate storage. IAM supports deploying server certificates in all regions, but you must obtain your certificate from an external provider for use with AWS.

You cannot upload an ACM certificate to IAM. Additionally, you cannot manage your certificates from the IAM Console. Use IAM as a certificate manager only when you need HTTPS connections in a region that is not supported by ACM.

Removing expired SSL/TLS certificates eliminates the risk that an invalid certificate will be deployed accidentally to a resource such as AWS Elastic Load Balancer (ELB), which can damage the credibility of the application/website behind the ELB. As a best practice, it is recommended to delete expired certificates.

Also have to update configurations at respective services to ensure there is no interruption in application/website access.

## Remediation

### From Command Line

Run the following command to delete the expired certificate:
```bash
aws iam delete-server-certificate --server-certificate-name <CERTIFICATE_NAME>
```
When the preceding command is successful, it does not return any output.

**Note**: By default, expired certificates never get deleted.




# CIS v1.3.0

| | TOTAL | 55 |
|-|-|-|
| ✅ | OK | 40 |
| ⇨ | Skip | 2 |
| ℹ | Info | 3 |
| ❌ | Alarm | 9 |
| ❗ | Error | 1 |

# 1 Identity and Access Management

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 16 | 0 | 0 | 5 | 1 | 22 |


## 1.1 Maintain current contact details
 
*Ensure contact email and telephone details for AWS accounts are current and map to more than one individual in your organization.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 0 | 0 | 0 | 1 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ❌ | is pretty insecure| `partition 10000` `us-east-2` `3335354343537`  |

## 1.2 Ensure security contact information is registered
 
*AWS provides customers with the option of specifying the contact information for accounts security team. It is recommended that this information be provided.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 0 | 0 | 0 | 1 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ❌ | is pretty insecure| `partition 10000` `us-east-2` `3335354343537`  |

## 1.3 Ensure security questions are registered in the AWS account
 
*The AWS support portal allows account owners to establish security questions that can be used to authenticate individuals calling AWS customer service for support. It is recommended that security questions be established.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 1 | 0 | 0 | 0 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ✅ | is totally secure and this is qa very very very very very long reason| `partition 30000` `us-east-3` `21323354377537`  |

## 1.4 Ensure no root user account access key exists
 
*The root user account is the most privileged user in an AWS account. AWS Access Keys provide programmatic access to a given AWS account. It is recommended that all access keys associated with the root user account be removed.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 1 | 0 | 0 | 0 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ✅ | is totally secure and this is qa very very very very very long reason| `partition 30000` `us-east-3` `21323354377537`  |

## 1.5 Ensure MFA is enabled for the "root user" account
 
*The root user account is the most privileged user in an AWS account. Multi-factor Authentication (MFA) adds an extra layer of protection on top of a username and password. With MFA enabled, when a user signs in to an AWS website, they will be prompted for their username and password as well as for an authentication code from their AWS MFA device.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 0 | 0 | 0 | 1 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ❌ | is pretty insecure| `partition 10000` `us-east-2` `3335354343537`  |

## 1.6 Ensure hardware MFA is enabled for the "root user" account
 
*The root user account is the most privileged user in an AWS account. MFA adds an extra layer of protection on top of a user name and password. With MFA enabled, when a user signs in to an AWS website, they will be prompted for their user name and password as well as for an authentication code from their AWS MFA device. For Level 2, it is recommended that the root user account be protected with a hardware MFA.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 0 | 0 | 0 | 0 | 1 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ❗ | is in some sort of error state| `partition 20000` `us-east-2` `21323354343537`  |

## 1.7 Eliminate use of the root user for administrative and daily tasks
 
*With the creation of an AWS account, a root user is created that cannot be disabled or deleted. That user has unrestricted access to and control over all resources in the AWS account. It is highly recommended that the use of this account be avoided for everyday tasks.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 1 | 0 | 0 | 0 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ✅ | is totally secure and this is qa very very very very very long reason| `partition 30000` `us-east-3` `21323354377537`  |

## 1.8 Ensure IAM password policy requires minimum length of 14 or greater
 
*Password policies are, in part, used to enforce password complexity requirements. IAM password policies can be used to ensure password are at least a given length. It is recommended that the password policy require a minimum password length 14.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 1 | 0 | 0 | 0 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ✅ | is totally secure and this is qa very very very very very long reason| `partition 30000` `us-east-3` `21323354377537`  |

## 1.9 Ensure IAM password policy prevents password reuse
 
*IAM password policies can prevent the reuse of a given password by the same user. It is recommended that the password policy prevent the reuse of passwords.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 1 | 0 | 0 | 0 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ✅ | is totally secure and this is qa very very very very very long reason| `partition 30000` `us-east-3` `21323354377537`  |

## 1.10 Ensure multi-factor authentication (MFA) is enabled for all IAM users that have a console password
 
*Multi-Factor Authentication (MFA) adds an extra layer of authentication assurance beyond traditional credentials. With MFA enabled, when a user signs in to the AWS Console, they will be prompted for their user name and password as well as for an authentication code from their physical or virtual MFA token. It is recommended that MFA be enabled for all accounts that have a console password.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 1 | 0 | 0 | 0 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ✅ | is totally secure and this is qa very very very very very long reason| `partition 30000` `us-east-3` `21323354377537`  |

## 1.11 Do not setup access keys during initial user setup for all IAM users that have a console password
 
*AWS console defaults to no check boxes selected when creating a new IAM user. When cerating the IAM User credentials you have to determine what type of access they require.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 1 | 0 | 0 | 0 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ✅ | is totally secure and this is qa very very very very very long reason| `partition 30000` `us-east-3` `21323354377537`  |

## 1.12 Ensure credentials unused for 90 days or greater are disabled
 
*AWS IAM users can access AWS resources using different types of credentials, such as passwords or access keys. It is recommended that all credentials that have been unused in 90 or greater days be deactivated or removed.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 1 | 0 | 0 | 0 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ✅ | is totally secure and this is qa very very very very very long reason| `partition 30000` `us-east-3` `21323354377537`  |

## 1.13 Ensure there is only one active access key available for any single IAM user
 
*Access keys are long-term credentials for an IAM user or the AWS account root user. You can use access keys to sign programmatic requests to the AWS CLI or AWS API. One of the best ways to protect your account is to not allow users to have multiple access keys.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 1 | 0 | 0 | 0 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ✅ | is totally secure and this is qa very very very very very long reason| `partition 30000` `us-east-3` `21323354377537`  |

## 1.14 Ensure access keys are rotated every 90 days or less
 
*Access keys consist of an access key ID and secret access key, which are used to sign programmatic requests that you make to AWS. AWS users need their own access keys to make programmatic calls to AWS from the AWS Command Line Interface (AWS CLI), Tools for Windows PowerShell, the AWS SDKs, or direct HTTP calls using the APIs for individual AWS services. It is recommended that all access keys be regularly rotated.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 1 | 0 | 0 | 0 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ✅ | is totally secure and this is qa very very very very very long reason| `partition 30000` `us-east-3` `21323354377537`  |

## 1.15 Ensure IAM Users Receive Permissions Only Through Groups
 
*IAM users are granted access to services, functions, and data through IAM policies. There are three ways to define policies for a user: 1) Edit the user policy directly, aka an inline, or user, policy; 2) attach a policy directly to a user; 3) add the user to an IAM group that has an attached policy.  Only the third implementation is recommended.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 0 | 0 | 0 | 1 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ❌ | is pretty insecure| `partition 10000` `us-east-2` `3335354343537`  |

## 1.16 Ensure IAM policies that allow full "*:*" administrative privileges are not attached
 
*IAM policies are the means by which privileges are granted to users, groups, or roles. It is recommended and considered a standard security advice to grant least privilege -that is, granting only the permissions required to perform a task. Determine what users need to do and then craft policies for them that let the users perform only those tasks, instead of allowing full administrative privileges.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 1 | 0 | 0 | 0 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ✅ | is totally secure and this is qa very very very very very long reason| `partition 30000` `us-east-3` `21323354377537`  |

## 1.17 Ensure a support role has been created to manage incidents with AWS Support
 
*AWS provides a support center that can be used for incident notification and response, as well as technical support and customer services. Create an IAM Role to allow authorized users to manage incidents with AWS Support.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 1 | 0 | 0 | 0 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ✅ | is totally secure and this is qa very very very very very long reason| `partition 30000` `us-east-3` `21323354377537`  |

## 1.18 Ensure IAM instance roles are used for AWS resource access from instances
 
*AWS access from within AWS instances can be done by either encoding AWS keys into AWS API calls or by assigning the instance to a role which has an appropriate permissions policy for the required access. "AWS Access" means accessing the APIs of AWS in order to access AWS resources or manage AWS account resources.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 1 | 0 | 0 | 0 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ✅ | is totally secure and this is qa very very very very very long reason| `partition 30000` `us-east-3` `21323354377537`  |

## 1.19 Ensure that all the expired SSL/TLS certificates stored in AWS IAM are removed
 
*To enable HTTPS connections to your website or application in AWS, you need an SSL/TLS server certificate. You can use ACM or IAM to store and deploy server certificates. Use IAM as a certificate manager only when you must support HTTPS connections in a region that is not supported by ACM. IAM securely encrypts your private keys and stores the encrypted version in IAM SSL certificate storage. IAM supports deploying server certificates in all regions, but you must obtain your certificate from an external provider for use with AWS. You cannot upload an ACM certificate to IAM. Additionally, you cannot manage your certificates from the IAM Console.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 1 | 0 | 0 | 0 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ✅ | is totally secure and this is qa very very very very very long reason| `partition 30000` `us-east-3` `21323354377537`  |

## 1.20 Ensure that S3 Buckets are configured with 'Block public access (bucket settings)'
 
*Amazon S3 provides Block public access (bucket settings) and Block public access (account settings) to help you manage public access to Amazon S3 resources. By default, S3 buckets and objects are created with public access disabled. However, an IAM principle with sufficient S3 permissions can enable public access at the bucket and/or object level. While enabled, Block public access (bucket settings) prevents an individual bucket, and its contained objects, from becoming publicly accessible. Similarly, Block public access (account settings) prevents all buckets, and contained objects, from becoming publicly accessible across the entire account.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 1 | 0 | 0 | 0 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ✅ | is totally secure and this is qa very very very very very long reason| `partition 30000` `us-east-3` `21323354377537`  |

## 1.21 Ensure that IAM Access analyzer is enabled
 
*Enable IAM Access analyzer for IAM policies about all resources. IAM Access Analyzer is a technology introduced at AWS reinvent 2019. After the Analyzer is enabled in IAM, scan results are displayed on the console showing the accessible resources. Scans show resources that other accounts and federated users can access, such as KMS keys and IAM roles. So the results allow you to determine if an unintended user is allowed, making it easier for administrators to monitor least privileges access.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 0 | 0 | 0 | 1 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ❌ | is pretty insecure| `partition 10000` `us-east-2` `3335354343537`  |

## 1.22 Ensure IAM users are managed centrally via identity federation or AWS Organizations for multi-account environments
 
*In multi-account environments, IAM user centralization facilitates greater user control. User access beyond the initial account is then provide via role assumption. Centralization of users can be accomplished through federation with an external identity provider or through the use of AWS Organizations.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 1 | 0 | 0 | 0 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ✅ | is totally secure and this is qa very very very very very long reason| `partition 30000` `us-east-3` `21323354377537`  |

# 2 Storage

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 2 | 0 | 1 | 0 | 0 | 3 |

# 2.1 Simple Storage Service (S3)

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 1 | 0 | 1 | 0 | 0 | 2 |


## 2.1.1 Ensure all S3 buckets employ encryption-at-rest
 
*Amazon S3 provides a variety of no, or low, cost encryption options to protect data at rest.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 1 | 0 | 0 | 0 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ✅ | is totally secure and this is qa very very very very very long reason| `partition 30000` `us-east-3` `21323354377537`  |

## 2.1.2 Ensure S3 Bucket Policy allows HTTPS requests
 
*At the Amazon S3 bucket level, you can configure permissions through a bucket policy making the objects accessible only through HTTPS.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 0 | 0 | 1 | 0 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ℹ | just some info, thought you should know| `partition 20000` `us-east-3` `21323354377537`  |

# 2.2 Elastic Compute Cloud (EC2)

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 1 | 0 | 0 | 0 | 0 | 1 |


## 2.2.1 Ensure EBS volume encryption is enabled
 
*Elastic Compute Cloud (EC2) supports encryption at rest when using the Elastic Block Store (EBS) service. While disabled by default, forcing encryption at EBS volume creation is supported.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 1 | 0 | 0 | 0 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ✅ | is totally secure and this is qa very very very very very long reason| `partition 30000` `us-east-3` `21323354377537`  |

# 3 Logging

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 11 | 0 | 0 | 0 | 0 | 11 |


## 3.1 Ensure CloudTrail is enabled in all regions
 
*AWS CloudTrail is a web service that records AWS API calls for your account and delivers log files to you. The recorded information includes the identity of the API caller, the time of the API call, the source IP address of the API caller, the request parameters, and the response elements returned by the AWS service. CloudTrail provides a history of AWS API calls for an account, including API calls made via the Management Console, SDKs, command line tools, and higher-level AWS services (such as CloudFormation).*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 1 | 0 | 0 | 0 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ✅ | is totally secure and this is qa very very very very very long reason| `partition 30000` `us-east-3` `21323354377537`  |

## 3.2 Ensure CloudTrail log file validation is enabled.
 
*CloudTrail log file validation creates a digitally signed digest file containing a hash of each log that CloudTrail writes to S3. These digest files can be used to determine whether a log file was changed, deleted, or unchanged after CloudTrail delivered the log. It is recommended that file validation be enabled on all CloudTrails.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 1 | 0 | 0 | 0 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ✅ | is totally secure and this is qa very very very very very long reason| `partition 30000` `us-east-3` `21323354377537`  |

## 3.3 Ensure the S3 bucket used to store CloudTrail logs is not publicly accessible
 
*CloudTrail logs a record of every API call made in your AWS account. These logs file are stored in an S3 bucket. It is recommended that the bucket policy or access control list (ACL) applied to the S3 bucket that CloudTrail logs to prevent public access to the CloudTrail logs.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 1 | 0 | 0 | 0 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ✅ | is totally secure and this is qa very very very very very long reason| `partition 30000` `us-east-3` `21323354377537`  |

## 3.4 Ensure CloudTrail trails are integrated with CloudWatch Logs
 
*AWS CloudTrail is a web service that records AWS API calls made in a given AWS account. The recorded information includes the identity of the API caller, the time of the API call, the source IP address of the API caller, the request parameters, and the response elements returned by the AWS service. CloudTrail uses Amazon S3 for log file storage and delivery, so log files are stored durably. In addition to capturing CloudTrail logs within a specified S3 bucket for long term analysis, realtime analysis can be performed by configuring CloudTrail to send logs to CloudWatch Logs. For a trail that is enabled in all regions in an account, CloudTrail sends log files from all those regions to a CloudWatch Logs log group. It is recommended that CloudTrail logs be sent to CloudWatch Logs.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 1 | 0 | 0 | 0 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ✅ | is totally secure and this is qa very very very very very long reason| `partition 30000` `us-east-3` `21323354377537`  |

## 3.5 Ensure AWS Config is enabled in all regions
 
*AWS Config is a web service that performs configuration management of supported AWS resources within your account and delivers log files to you. The recorded information includes the configuration item (AWS resource), relationships between configuration items (AWS resources), any configuration changes between resources. It is recommended to enable AWS Config be enabled in all regions.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 1 | 0 | 0 | 0 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ✅ | is totally secure and this is qa very very very very very long reason| `partition 30000` `us-east-3` `21323354377537`  |

## 3.6 Ensure S3 bucket access logging is enabled on the CloudTrail S3 bucket
 
*S3 Bucket Access Logging generates a log that contains access records for each request made to your S3 bucket. An access log record contains details about the request, such as the request type, the resources specified in the request worked, and the time and date the request was processed. It is recommended that bucket access logging be enabled on the CloudTrail S3 bucket.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 1 | 0 | 0 | 0 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ✅ | is totally secure and this is qa very very very very very long reason| `partition 30000` `us-east-3` `21323354377537`  |

## 3.7 Ensure CloudTrail logs are encrypted at rest using KMS CMKs
 
*AWS CloudTrail is a web service that records AWS API calls for an account and makes those logs available to users and resources in accordance with IAM policies. AWS Key Management Service (KMS) is a managed service that helps create and control the encryption keys used to encrypt account data, and uses Hardware Security Modules (HSMs) to protect the security of encryption keys. CloudTrail logs can be configured to leverage server side encryption (SSE) and KMS customer created master keys (CMK) to further protect CloudTrail logs. It is recommended that CloudTrail be configured to use SSE-KMS.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 1 | 0 | 0 | 0 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ✅ | is totally secure and this is qa very very very very very long reason| `partition 30000` `us-east-3` `21323354377537`  |

## 3.8 Ensure rotation for customer created CMKs is enabled
 
*AWS Key Management Service (KMS) allows customers to rotate the backing key which is key material stored within the KMS which is tied to the key ID of the Customer Created customer master key (CMK). It is the backing key that is used to perform cryptographic operations such as encryption and decryption. Automated key rotation currently retains all prior backing keys so that decryption of encrypted data can take place transparently. It is recommended that CMK key rotation be enabled.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 1 | 0 | 0 | 0 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ✅ | is totally secure and this is qa very very very very very long reason| `partition 30000` `us-east-3` `21323354377537`  |

## 3.9 Ensure VPC flow logging is enabled in all VPCs
 
*VPC Flow Logs is a feature that enables you to capture information about the IP traffic going to and from network interfaces in your VPC. After you've created a flow log, you can view and retrieve its data in Amazon CloudWatch Logs. It is recommended that VPC Flow Logs be enabled for packet "Rejects" for VPCs.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 1 | 0 | 0 | 0 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ✅ | is totally secure and this is qa very very very very very long reason| `partition 30000` `us-east-3` `21323354377537`  |

## 3.10 Ensure that Object-level logging for write events is enabled for S3 bucket
 
*S3 object-level API operations such as GetObject, DeleteObject, and PutObject are called data events. By default, CloudTrail trails don't log data events and so it is recommended to enable Object-level logging for S3 buckets.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 1 | 0 | 0 | 0 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ✅ | is totally secure and this is qa very very very very very long reason| `partition 30000` `us-east-3` `21323354377537`  |

## 3.11 Ensure that Object-level logging for read events is enabled for S3 bucket
 
*S3 object-level API operations such as GetObject, DeleteObject, and PutObject are called data events. By default, CloudTrail trails don't log data events and so it is recommended to enable Object-level logging for S3 buckets.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 1 | 0 | 0 | 0 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ✅ | is totally secure and this is qa very very very very very long reason| `partition 30000` `us-east-3` `21323354377537`  |

# 4 Monitoring

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 11 | 2 | 2 | 0 | 0 | 15 |


## 4.1 Ensure a log metric filter and alarm exist for unauthorized API calls
 
*Real-time monitoring of API calls can be achieved by directing CloudTrail Logs to CloudWatch Logs and establishing corresponding metric filters and alarms. It is recommended that a metric filter and alarm be established for unauthorized API calls.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 1 | 0 | 0 | 0 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ✅ | is totally secure and this is qa very very very very very long reason| `partition 30000` `us-east-3` `21323354377537`  |

## 4.2 Ensure a log metric filter and alarm exist for Management Console sign-in without MFA
 
*Real-time monitoring of API calls can be achieved by directing CloudTrail Logs to CloudWatch Logs and establishing corresponding metric filters and alarms. It is recommended that a metric filter and alarm be established for console logins that are not protected by multi-factor authentication (MFA).*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 1 | 0 | 0 | 0 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ✅ | is totally secure and this is qa very very very very very long reason| `partition 30000` `us-east-3` `21323354377537`  |

## 4.3 Ensure a log metric filter and alarm exist for usage of "root" account
 
*Real-time monitoring of API calls can be achieved by directing CloudTrail Logs to CloudWatch Logs and establishing corresponding metric filters and alarms. It is recommended that a metric filter and alarm be established for root login attempts.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 0 | 0 | 1 | 0 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ℹ | just some info, thought you should know| `partition 20000` `us-east-3` `21323354377537`  |

## 4.4 Ensure a log metric filter and alarm exist for IAM policy changes
 
*Real-time monitoring of API calls can be achieved by directing CloudTrail Logs to CloudWatch Logs and establishing corresponding metric filters and alarms. It is recommended that a metric filter and alarm be established changes made to Identity and Access Management (IAM) policies.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 1 | 0 | 0 | 0 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ✅ | is totally secure and this is qa very very very very very long reason| `partition 30000` `us-east-3` `21323354377537`  |

## 4.5 Ensure a log metric filter and alarm exist for CloudTrail configuration changes
 
*Real-time monitoring of API calls can be achieved by directing CloudTrail Logs to CloudWatch Logs and establishing corresponding metric filters and alarms. It is recommended that a metric filter and alarm be established for detecting changes to CloudTrail's configurations.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 1 | 0 | 0 | 0 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ✅ | is totally secure and this is qa very very very very very long reason| `partition 30000` `us-east-3` `21323354377537`  |

## 4.6 Ensure a log metric filter and alarm exist for AWS Management Console authentication failures
 
*Real-time monitoring of API calls can be achieved by directing CloudTrail Logs to CloudWatch Logs and establishing corresponding metric filters and alarms. It is recommended that a metric filter and alarm be established for failed console authentication attempts.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 0 | 0 | 1 | 0 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ℹ | just some info, thought you should know| `partition 20000` `us-east-3` `21323354377537`  |

## 4.7 Ensure a log metric filter and alarm exist for disabling or scheduled deletion of customer created CMKs
 
*Real-time monitoring of API calls can be achieved by directing CloudTrail Logs to CloudWatch Logs and establishing corresponding metric filters and alarms. It is recommended that a metric filter and alarm be established for customer created CMKs which have changed state to disabled or scheduled deletion.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 1 | 0 | 0 | 0 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ✅ | is totally secure and this is qa very very very very very long reason| `partition 30000` `us-east-3` `21323354377537`  |

## 4.8 Ensure a log metric filter and alarm exist for S3 bucket policy changes
 
*Real-time monitoring of API calls can be achieved by directing CloudTrail Logs to CloudWatch Logs and establishing corresponding metric filters and alarms. It is recommended that a metric filter and alarm be established for changes to S3 bucket policies.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 1 | 0 | 0 | 0 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ✅ | is totally secure and this is qa very very very very very long reason| `partition 30000` `us-east-3` `21323354377537`  |

## 4.9 Ensure a log metric filter and alarm exist for AWS Config configuration changes
 
*Real-time monitoring of API calls can be achieved by directing CloudTrail Logs to CloudWatch Logs and establishing corresponding metric filters and alarms. It is recommended that a metric filter and alarm be established for detecting changes to CloudTrail's configurations.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 0 | 1 | 0 | 0 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ⇨ | totally skipping this one| `partition 40000` `us-east-4` `21323354377537`  |

## 4.10 Ensure a log metric filter and alarm exist for security group changes
 
*Real-time monitoring of API calls can be achieved by directing CloudTrail Logs to CloudWatch Logs and establishing corresponding metric filters and alarms. Security Groups are a stateful packet filter that controls ingress and egress traffic within a VPC. It is recommended that a metric filter and alarm be established for detecting changes to Security Groups.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 1 | 0 | 0 | 0 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ✅ | is totally secure and this is qa very very very very very long reason| `partition 30000` `us-east-3` `21323354377537`  |

## 4.11 Ensure a log metric filter and alarm exist for changes to Network Access Control Lists (NACL)
 
*Real-time monitoring of API calls can be achieved by directing CloudTrail Logs to CloudWatch Logs and establishing corresponding metric filters and alarms. NACLs are used as a stateless packet filter to control ingress and egress traffic for subnets within a VPC. It is recommended that a metric filter and alarm be established for changes made to NACLs.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 1 | 0 | 0 | 0 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ✅ | is totally secure and this is qa very very very very very long reason| `partition 30000` `us-east-3` `21323354377537`  |

## 4.12 Ensure a log metric filter and alarm exist for changes to network gateways
 
*Real-time monitoring of API calls can be achieved by directing CloudTrail Logs to CloudWatch Logs and establishing corresponding metric filters and alarms. Network gateways are required to send/receive traffic to a destination outside of a VPC. It is recommended that a metric filter and alarm be established for changes to network gateways.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 1 | 0 | 0 | 0 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ✅ | is totally secure and this is qa very very very very very long reason| `partition 30000` `us-east-3` `21323354377537`  |

## 4.13 Ensure a log metric filter and alarm exist for route table changes
 
*Real-time monitoring of API calls can be achieved by directing CloudTrail Logs to CloudWatch Logs and establishing corresponding metric filters and alarms. Routing tables are used to route network traffic between subnets and to network gateways. It is recommended that a metric filter and alarm be established for changes to route tables.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 1 | 0 | 0 | 0 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ✅ | is totally secure and this is qa very very very very very long reason| `partition 30000` `us-east-3` `21323354377537`  |

## 4.14 Ensure a log metric filter and alarm exist for VPC changes
 
*Real-time monitoring of API calls can be achieved by directing CloudTrail Logs to CloudWatch Logs and establishing corresponding metric filters and alarms. It is possible to have more than 1 VPC within an account, in addition it is also possible to create a peer connection between 2 VPCs enabling network traffic to route between VPCs. It is recommended that a metric filter and alarm be established for changes made to VPCs.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 0 | 1 | 0 | 0 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ⇨ | totally skipping this one| `partition 40000` `us-east-4` `21323354377537`  |

## 4.15 Ensure a log metric filter and alarm exists for AWS Organizations changes
 
*Real-time monitoring of API calls can be achieved by directing CloudTrail Logs to CloudWatch Logs and establishing corresponding metric filters and alarms. It is recommended that a metric filter and alarm be established for AWS Organizations changes made in the master AWS Account.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 1 | 0 | 0 | 0 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ✅ | is totally secure and this is qa very very very very very long reason| `partition 30000` `us-east-3` `21323354377537`  |

# 5 Networking

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 0 | 0 | 0 | 4 | 0 | 4 |


## 5.1 Ensure no Network ACLs allow ingress from 0.0.0.0/0 to remote server administration ports
 
*The Network Access Control List (NACL) function provide stateless filtering of ingress and egress network traffic to AWS resources. It is recommended that no NACL allows unrestricted ingress access to remote server administration ports, such as SSH to port 22 and RDP to port 3389.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 0 | 0 | 0 | 1 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ❌ | is pretty insecure| `partition 10000` `us-east-2` `3335354343537`  |

## 5.2 Ensure no security groups allow ingress from 0.0.0.0/0 to remote server administration ports
 
*Security groups provide stateful filtering of ingress and egress network traffic to AWS resources. It is recommended that no security group allows unrestricted ingress access to remote server administration ports, such as SSH to port 22 and RDP to port 3389.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 0 | 0 | 0 | 1 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ❌ | is pretty insecure| `partition 10000` `us-east-2` `3335354343537`  |

## 5.3 Ensure the default security group of every VPC restricts all traffic
 
*A VPC comes with a default security group whose initial settings deny all inbound traffic, allow all outbound traffic, and allow all traffic between instances assigned to the security group. If you don't specify a security group when you launch an instance, the instance is automatically assigned to this default security group. Security groups provide stateful filtering of ingress/egress network traffic to AWS resources. It is recommended that the default security group restrict all traffic.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 0 | 0 | 0 | 1 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ❌ | is pretty insecure| `partition 10000` `us-east-2` `3335354343537`  |

## 5.4 Ensure routing tables for VPC peering are 'least access'
 
*A VPC comes with a default security group whose initial settings deny all inbound traffic, allow all outbound traffic, and allow all traffic between instances assigned to the security group. If you don't specify a security group when you launch an instance, the instance is automatically assigned to this default security group. Security groups provide stateful filtering of ingress/egress network traffic to AWS resources. It is recommended that the default security group restrict all traffic.*

| OK | Skip | Info | Alarm | Error | Total |
|-|-|-|-|-|-|
| 0 | 0 | 0 | 1 | 0 | 1 |



| | Reason | Dimensions |
|-|--------|------------|
| ❌ | is pretty insecure| `partition 10000` `us-east-2` `3335354343537`  |






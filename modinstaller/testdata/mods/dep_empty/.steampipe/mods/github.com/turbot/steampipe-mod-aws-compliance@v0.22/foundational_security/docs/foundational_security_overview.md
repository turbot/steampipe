## Overview

The AWS Foundational Security Best Practices standard is a set of controls that
detect when your deployed accounts and resources deviate from security best
practices.

The standard allows you to continuously evaluate all of your AWS accounts and
workloads to quickly identify areas of deviation from best practices. It
provides actionable and prescriptive guidance on how to improve and maintain
your organization’s security posture.

## Control Categories

These are the available categories for AWS Security Hub controls. The category
for a control reflects the security function that the control applies to.

### Identify

Develop the organizational understanding to manage cybersecurity risk to
systems, assets, data, and capabilities.

**Inventory**
- Has the service implemented the correct resource tagging strategies? Do the tagging strategies include the resource owner?
- What resources does the service use? Are they approved resources for this service?
- Do you have visibility into the approved inventory? For example, do you use services such as Amazon EC2 Systems Manager and AWS Service Catalog?

**Logging**
Have you securely enabled all relevant logging for the service? Examples of log files include the following:

- Amazon VPC Flow Logs
- Elastic Load Balancing access logs
- Amazon CloudFront logs
- Amazon CloudWatch Logs
- Amazon Relational Database Service logging
- Amazon Elasticsearch Service slow index logs
- X-Ray tracing
- AWS Directory Service logs
- AWS Config items
- Snapshots

### Protect

Develop and implement the appropriate safeguards to ensure delivery of critical
infrastructure services and secure coding practices.

**Secure access management**
- Does the service use least privilege practices in its IAM or resource policies?
- Are passwords and secrets sufficiently complex? Are they rotated appropriately?
- Does the service use multi-factor authentication (MFA)?
- Does the service avoid the root account?
- Do resource-based policies allow public access?

**Secure network configuration**
- Does the service avoid public and insecure remote network access?
- Does the service use VPCs properly? For example, are jobs required to run in VPCs?
- Does the service properly segment and isolate sensitive resources?

**Data protection**
- Encryption of data at rest – Does the service encrypt data at rest?
- Encryption of data in transit – Does the service encrypt data in transit?
- Data integrity – Does the service validate data for integrity?
- Data deletion protection – Does the service protect data from accidental deletion?
- Data management / usage – Do you use services such as Amazon Macie to track the location of your sensitive data?

**API protection**
- Does the service use AWS PrivateLink to protect the service API operations?

**Protective services**
- Are the correct protective services in place? Do they provide the correct amount of coverage?
- Protective services help you deflect attacks and compromises that are directed at the service. Examples of protective services in AWS include AWS Control Tower, AWS WAF, AWS Shield Advanced, Vanta, Secrets Manager, IAM Access Analyzer, and AWS Resource Access Manager.

**Secure development**
- Do you use secure coding practices?
- Do you avoid vulnerabilities such as the Open Web Application Security Project (OWASP) Top Ten?

### Detect

Develop and implement the appropriate activities to identify the occurrence of
a cybersecurity event.

**Detection services**
- Are the correct detection services in place?
- Do they provide the correct amount of coverage?
- Examples of AWS detection services include Amazon GuardDuty, AWS Security Hub, Amazon Inspector, Amazon Detective, Amazon CloudWatch Alarms, AWS IoT Device Defender, and AWS Trusted Advisor.

### Respond

Develop and implement the appropriate activities to take action regarding a
detected cybersecurity event.

**Response actions**
- Do you respond to security events swiftly?
- Do you have any active critical or high severity findings?

**Forensics**
- Can you securely acquire forensic data for the service? For example, do you acquire Amazon EBS snapshots associated with true positive findings?
- Have you set up a forensic account?

### Recover

Develop and implement the appropriate activities to maintain plans for
resilience and to restore any capabilities or services that were impaired due
to a cybersecurity event.

**Resilience**
- Does the service configuration support graceful failovers, elastic scaling, and high availability?
- Have you established backups?

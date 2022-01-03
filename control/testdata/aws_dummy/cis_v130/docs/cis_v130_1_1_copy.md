## Description
Ensure your *Contact Information* and *Alternate Contacts* are correct in the AWS account settings page of your AWS account.  

In addition to the primary contact information, you may enter the following contacts:
- **Billing**: When your monthly invoice is available, or your payment method needs to be updated. If your Receive PDF Invoice By Email is turned on in your Billing preferences, your alternate billing contact will receive the PDF invoices as well.
- **Operations**: When your service is, or will be, temporarily unavailable in one of more Regions. Any notification related to operations.
- **Security**:  When you have notifications from the AWS Abuse team for potentially fraudulent activity on your AWS account. Any notification related to security.

As a best practice, avoid using contact information for individuals, and instead use group email addresses and shared company phone numbers.

## Rationale
AWS uses the contact information to inform you of important service events, billing issues, and security issues.  Keeping your contact information up to date ensure timely delivery of important information to the relevant stakeholders.  Incorrect contact information may result in communications delays that could impact your ability to operate.  


## Remediation
There is no API available for setting contact information - you must log in to the AWS console to verify and set your contact information.  

1. Sign into the AWS console, and navigate to the [Account Settings](https://console.aws.amazon.com/billing/home?#/account) page.
1. Verify that the information in the **Contact Information** section is correct and complete.  If changes are required, click **Edit**, make your changes, and then click **Update**.
1. Verify that the information in the **Alternate Contacts** section is correct and complete.  If changes are required, click **Edit**, make your changes, and then click **Update**.
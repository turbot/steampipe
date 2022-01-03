## Description

Elastic Compute Cloud (EC2) supports encryption at rest when using the Elastic Block Store(EBS) service. While disabled by default, forcing encryption at EBS volume creation is supported.

Default EBS volume encryption only applies to newly created EBS volumes. Existing EBS volumes are not converted automatically.

Encrypting data at rest reduces the likelihood that it is unintentionally exposed and can nullify the impact of disclosure if the encryption remains unbroken.

## Remediation

### From Console

1. Open the Amazon EC2 console using [EC2](https://console.aws.amazon.com/ec2/)
2. Under **Account attributes**, click `EBS encryption`.
3. Click **Manage**.
4. Click the `Enable` checkbox.
5. Click `Update EBS encryption`
6. Repeat for every region requiring the change.

### From Command Line

1. Run
```bash
aws --region <region> ec2 enable-ebs-encryption-by-default.
```
2. Verify that **EbsEncryptionByDefault**: **true** is displayed.
3. Review every region in-use.

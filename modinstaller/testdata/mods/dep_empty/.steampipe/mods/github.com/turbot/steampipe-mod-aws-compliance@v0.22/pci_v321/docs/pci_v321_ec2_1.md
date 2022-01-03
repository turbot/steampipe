## Description

This control checks whether Amazon Elastic Block Store snapshots are not publicly restorable by everyone, which makes them public. Amazon EBS snapshots should not be publicly restorable by everyone unless you explicitly allow it, to avoid accidental exposure of your company’s sensitive data.

You should also ensure that permission to change Amazon EBS configurations are restricted to authorized AWS accounts only. Learn more about managing Amazon EBS snapshot permissions in the [Amazon EC2 User Guide for Linux Instances](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ebs-modifying-snapshot-permissions.html).

## Remediation

1. Open the [Amazon EC2 console](https://console.aws.amazon.com/ec2/).
2. In the navigation pane, under `Elastic Block Store`, choose **Snapshots** and then select your public snapshot.
3. Choose **Permissions** tab from metadata view section, click **Edit**
4. Choose **Private** in `This snapshot is currently`
5. Choose **Save**

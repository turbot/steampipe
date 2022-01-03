## Description

This control checks whether the compliance status of the Amazon EC2 Systems Manager patch compliance is `COMPLIANT` or `NON_COMPLIANT` after the patch installation on the instance. It only checks instances that are managed by Systems Manager Patch Manager.

Having your EC2 instances fully patched as required by your organization reduces the attack surface of your AWS accounts.

## Remediation

To remediate this issue, install the required patches on your noncompliant instances.

**To remediate noncompliant patches**

1. Open the [AWS Systems Manager console](https://console.aws.amazon.com/systems-manager/).
2. Under `Instances & Nodes`, choose `Run Command` and then choose `Run command`.
3. Choose the button next to `AWS-RunPatchBaseline`.
4. Change the `Operation` to `Install`.
5. Choose `Choose instances manually` and then choose the noncompliant instances.
6. At the bottom of the page, choose `Run`.
7. After the command is complete, to monitor the new compliance status of your patched instances, in the navigation pane, choose `Compliance`.
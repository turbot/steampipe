## Description

This control checks whether the compliance status of the Amazon EC2 Systems Manager patch compliance is COMPLIANT or NON_COMPLIANT after the patch installation on the instance.

It only checks instances that are managed by AWS Systems Manager Patch Manager.

It does not check whether the patch was applied within the 30-day limit prescribed by PCI DSS requirement 6.2.

It also does not validate whether the patches applied were classified as security patches.

## Remediation

This rule checks whether the compliance status of the Amazon EC2 Systems Manager patch compliance is COMPLIANT or NON_COMPLIANT. To find out more about patch compliance states, see the [AWS Systems Manager User Guide](https://docs.aws.amazon.com/systems-manager/latest/userguide/about-patch-compliance-states.html).

1. Open the [AWS Systems Manager console](https://console.aws.amazon.com/systems-manager/.)
2. In the navigation pane, under **Instances & Nodes**, choose **Run Command**.
3. Choose **Run command**.
4. Choose the radio button next to AWS-RunPatchBaseline and then change the **Operation to Install**.
5. Choose **Choose instances manually** and then choose the noncompliant instance(s).
6. Scroll to the bottom and then choose **Run**.
7. After the command has completed, to monitor the new compliance status of your patched instances, in the navigation pane, choose Compliance.

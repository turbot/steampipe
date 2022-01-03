## Description

This control checks whether the status of the AWS Systems Manager association compliance is COMPLIANT or NON_COMPLIANT after the association is run on an instance. The control passes if the association compliance status is COMPLIANT.

A State Manager association is a configuration that is assigned to your managed instances. The configuration defines the state that you want to maintain on your instances. For example, an association can specify that antivirus software must be installed and running on your instances, or that certain ports must be closed.

After you create one or more State Manager associations, compliance status information is immediately available to you in the console or in response to AWS CLI commands or corresponding Systems Manager API operations. For associations, Configuration Compliance shows statuses of Compliant or Non-compliant and the severity level assigned to the association, such as Critical or Medium. To learn more about State Manager association compliance, see About State Manager association compliance in the AWS Systems Manager User Guide.

You must configure your in-scope EC2 instances for Systems Manager association. You must also configure the patch baseline for the security rating of the vendor of patches, and set the autoapproval date to meet PCI DSS 3.2.1 requirement 6.2. For additional guidance on how to create an association, see [Create an association](https://docs.aws.amazon.com/systems-manager/latest/userguide/sysman-state-assoc.html) in the AWS Systems Manager User Guide. For additional information on working with patching in Systems Manager, see [AWS Systems Manager Patch Manager](https://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-patch.html) in the AWS Systems Manager User Guide.

## Remediation

A failed association can be related to different things, including targets and SSM document names. To remediate this issue, you must first identify and investigate the association. You can then update the association to correct the specific issue.

You can edit an association to specify a new name, schedule, severity level, or targets. After you edit an association, Systems Manager creates a new version. To investigate and update a failed association

1. Open the [AWS Systems Manager console](https://console.aws.amazon.com/systems-manager/).
2. In the navigation pane, under **Instances & Nodes**, choose **Managed Instances**.
3. Choose the instance ID that has an **Association status** of **Failed**.
4. Choose **View details**.
5. Choose **Associations**.
6. Note the name of the association that has an **Association status** of **Failed**. This is the association that you need to investigate. You need to use the association name in the next step.
7. In the navigation pane, choose **State Manager**. Search for the association name, then select the association.
8. After you determine the issue, edit the failed association to correct the problem. For information on how to edit an association, see [Edit an association](https://docs.aws.amazon.com/systems-manager/latest/userguide/sysman-state-assoc-edit.html).

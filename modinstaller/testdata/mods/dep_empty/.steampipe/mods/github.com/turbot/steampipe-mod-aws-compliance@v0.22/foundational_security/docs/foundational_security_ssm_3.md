## Description

This control checks whether the status of the AWS Systems Manager association compliance is `COMPLIANT` or `NON_COMPLIANT` after the association is run on an instance. The control passes if the association compliance status is `COMPLIANT`.

A State Manager association is a configuration that is assigned to your managed instances. The configuration defines the state that you want to maintain on your instances. For example, an association can specify that antivirus software must be installed and running on your instances or that certain ports must be closed.

After you create one or more State Manager associations, compliance status information is immediately available to you. You can view the compliance status in the console or in response to AWS CLI commands or corresponding Systems Manager API actions. For associations, Configuration Compliance shows the compliance status (`Compliant` or `Non-compliant`). It also shows the severity level assigned to the association, such as `Critical` or `Medium`.

## Remediation

A failed association can be related to different things, including targets and SSM document names. To remediate this issue, you must first identify and investigate the association. You can then update the association to correct the specific issue.

You can edit an association to specify a new name, schedule, severity level, or targets. After you edit an association, AWS Systems Manager creates a new version.

**To investigate and update a failed association**

1. Open the [AWS Systems Manager console](https://console.aws.amazon.com/systems-manager/).
2. In the navigation pane, under `Instances & Nodes`, choose `Managed Instances`.
3. Choose the instance ID that has an `Association status` of `Failed`.
4. Choose `View details`.
5. Choose `Associations`.
6. Note the name of the association that has an `Association status` of `Failed`. This is the association that you need to investigate. You need to use the association name in the next step.
7. In the navigation pane, under `Instances & Nodes,` choose `State Manager`. Search for the association name, then select the association.
8. After you determine the issue, edit the failed association to correct the problem. For information on how to edit an association, see [Edit an association](https://docs.aws.amazon.com/systems-manager/latest/userguide/sysman-state-assoc-edit.html).
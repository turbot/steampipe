## Description

This control checks whether an Amazon ECS task definition that has host networking mode also has 'privileged' or 'user' container definitions. The control fails for task definitions that have host network mode and container definitions where privileged=false or is empty and user=root or is empty.

If a task definition has elevated privileges, it is because the customer has specifically opted in to that configuration. This control checks for unexpected privilege escalation when a task definition has host networking enabled but the customer has not opted in to elevated privileges.

## Remediation

For information on how to update a task definition, see [Updating a task definition](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/update-task-definition.html).

Note that when you update a task definition, it does not update running tasks that were launched from the previous task definition. To update a running task, you must redeploy the task with the new task definition.
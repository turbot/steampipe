## Description

AWS access from AWS instances can be done by either encoding AWS keys into AWS API calls or by assigning the instance to a role which has an appropriate permissions policy for the required access. *AWS Access* means accessing the APIs of AWS in order to access AWS resources or manage AWS account resources.

AWS IAM roles reduce the risks associated with sharing and rotating credentials that can be used outside of AWS itself. If credentials are compromised, they can be used from outside of the the AWS account they give access to. In contrast, in order to leverage role permissions an attacker would need to gain and maintain access to a specific instance to use the privileges associated with it.

Additionally, if credentials are encoded into compiled applications or other hard to change mechanisms, then they are even more unlikely to be properly rotated due to service disruption risks. As time goes on, credentials that cannot be rotated are more likely to be known by an increasing number of individuals who no longer work for the organization owning the credentials.

## Remediation

### From Console

Perform the following action to check whether an Instance is associated with a role:

1. Sign into the AWS console as a user(with appropriate permissions to view identity access management account settings).
2. Open the EC2 Dashboard and choose **Instances**.
3. Click the EC2 instance that performs AWS actions, in the lower pane details you can find **IAM Role**.
4. If the IAM Role is blank, the instance is not assigned to any role.
5. If the Role is filled in, it does not mean the instance might not also have credentials encoded on it for some activities.
6. Audit all scripts and environment variables to ensure that none of them contain AWS credentials.
7. Also examine all source code and configuration files of the application to verify if there is any credentials stored.

**Note**: IAM roles can only be associated at the launch of an instance. To add a role to an instance, you must create a new instance.

Perform the following action to create and attach a role to an Instance:

1. In AWS IAM create a new role. Attach the right permissions policy as needed.
2. In the AWS console launch a new instance with identical settings to the existing instance, and ensure that the newly created role is selected in **Configure Instance Details** page.
3. Shutdown both the existing and the new instances.
4. Detach disks from both the instances.
5. Attach the existing instance disks to the new instance.
6. Boot the new instance and you should have the same machine, but with the associated role with right level of permissions.

**Note**:
- When your environment has dependencies on a dynamically assigned **PRIVATE IP** address, you can create an AMI from the existing instance, destroy the old one and then when launching from the AMI, manually assign the previous private IP address.
- When your environment has dependencies on a dynamically assigned **PUBLIC IP** address, ensure the address is retained and assign an instance role. Dependencies on dynamically assigned public IP addresses are a bad practice and, if possible, you may wish to rebuild the instance with a new elastic IP address and make the investment to remediate affected systems while assigning the system to a role.
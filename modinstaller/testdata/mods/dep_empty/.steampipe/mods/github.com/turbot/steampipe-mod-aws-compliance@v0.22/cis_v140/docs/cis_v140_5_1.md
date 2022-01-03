## Description

The Network Access Control List (NACL) function provide stateless filtering of ingress and egress network traffic to AWS resources. It is recommended that no NACL allows unrestricted ingress access to remote server administration ports, such as SSH to port 22
and RDP to port 3389.

Public access to remote server administration ports, such as 22 and 3389, increases resource attack surface and unnecessarily raises the risk of resource compromise.

## Remediation

### From Console

1. Login VPC Console at [VPC](https://console.aws.amazon.com/vpc/home)
2. In the left pane, click **Network ACLs**
3. For each network ACL to remediate, perform the following:
4. Select the respective network ACL
5. Choose the **Inbound Rules** tab and click **Edit Inbound rules**
6. Check remote server access port entries. **Note** A Port value of ALL or a port range such as 0-1024 are inclusive of port 22, 3389, and other remote server administration ports.
7. Either
    - A) Update the `source` field to a range other than 0.0.0.0/0 or
    - B) Click **Remove** the offending inbound rule
8. Click **Save changes**

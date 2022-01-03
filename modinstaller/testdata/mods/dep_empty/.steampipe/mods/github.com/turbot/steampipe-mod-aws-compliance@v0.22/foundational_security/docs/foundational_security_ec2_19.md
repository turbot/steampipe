## Description

This control checks whether unrestricted incoming traffic for the security groups is accessible to the specified ports that have the highest risk. This control passes when none of the rules in a security group allow ingress traffic from 0.0.0.0/0 for those ports.

Unrestricted access (0.0.0.0/0) increases opportunities for malicious activity, such as hacking, denial-of-service attacks, and loss of data.

Security groups provide stateful filtering of ingress and egress network traffic to AWS resources. No security group should allow unrestricted ingress access to the following ports:

- 3389 (RDP)

- 20, 21 (FTP)

- 22 (SSH)

- 23 (Telnet)

- 110 (POP3)

- 143 (IMAP)

- 3306 (mySQL)

- 8080 (proxy)

- 1433, 1434 (MSSQL)

- 9200 or 9300 (Elasticsearch)

- 5601 (Kibana)

- 25 (SMTP)

- 445 (CIFS)

- 135 (RPC)

- 4333 (ahsp)

- 5432 (postgresql)

- 5500 (fcp-addr-srvr1)

## Remediation

For information on how to delete rules from a security group, see [Delete rules from a security group](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/working-with-security-groups.html#deleting-security-group-rule) in the Amazon EC2 User Guide for Linux Instances.
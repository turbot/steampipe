## Description

Once a VPC peering connection is established, routing tables must be updated to establish any connections between the peered VPCs. These routes can be as specific as desired - even peering a VPC to only a single host on the other side of the connection.

Being highly selective in peering routing tables is a very effective way of minimizing the impact of breach as resources outside of these routes are inaccessible to the peered VPC.

## Remediation

Remove and add route table entries to ensure that the least number of subnets or hosts as is required to accomplish the purpose for peering are routable.

### From Command Line

1. For each <route_table_id> containing routes non compliant with your routing policy (which grants more than desired "least access"), delete the non compliant route:

```bash
aws ec2 delete-route --route-table-id <route_table_id> --destination-cidrblock <non_compliant_destination_CIDR>
```

2. Create a new compliant route:

```bash
aws ec2 create-route --route-table-id <route_table_id> --destination-cidrblock <compliant_destination_CIDR> --vpc-peering-connection-id <peering_connection_id>
```

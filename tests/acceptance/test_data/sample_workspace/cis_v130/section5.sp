locals {
  cis_v130_5_common_tags = merge(local.cis_v130_common_tags, {
    cis_section_id = "5"
  })
}

benchmark "cis_v130_5" {
  title = "5 Networking"
  #documentation = file("./cis_v130/docs/cis_v130_5.md")
  tags = local.cis_v130_5_common_tags
  children = [
    control.cis_v130_5_1,
    control.cis_v130_5_2,
    control.cis_v130_5_3,
    control.cis_v130_5_4
  ]
}

control "cis_v130_5_1" {
  title       = "5.1 Ensure no Network ACLs allow ingress from 0.0.0.0/0 to remote server administration ports"
  description = "The Network Access Control List (NACL) function provide stateless filtering of ingress and egress network traffic to AWS resources. It is recommended that no NACL allows unrestricted ingress access to remote server administration ports, such as SSH to port 22 and RDP to port 3389."
  sql         = query.alarm.sql
  #documentation = file("./cis_v130/docs/cis_v130_5_1.md")

  tags = merge(local.cis_v130_5_common_tags, {
    cis_item_id  = "5.1"
    cis_type     = "automated"
    cis_levels   = "1"
    cis_controls = "9.2,12.4"
  })
}

control "cis_v130_5_2" {
  title       = "5.2 Ensure no security groups allow ingress from 0.0.0.0/0 to remote server administration ports"
  description = "Security groups provide stateful filtering of ingress and egress network traffic to AWS resources. It is recommended that no security group allows unrestricted ingress access to remote server administration ports, such as SSH to port 22 and RDP to port 3389."
  sql         = query.alarm.sql
  #documentation = file("./cis_v130/docs/cis_v130_5_2.md")

  tags = merge(local.cis_v130_5_common_tags, {
    cis_item_id  = "5.2"
    cis_type     = "automated"
    cis_levels   = "1"
    cis_controls = "9.2,12.4"
  })
}

control "cis_v130_5_3" {
  title       = "5.3 Ensure the default security group of every VPC restricts all traffic"
  description = "A VPC comes with a default security group whose initial settings deny all inbound traffic, allow all outbound traffic, and allow all traffic between instances assigned to the security group. If you don't specify a security group when you launch an instance, the instance is automatically assigned to this default security group. Security groups provide stateful filtering of ingress/egress network traffic to AWS resources. It is recommended that the default security group restrict all traffic."
  sql         = query.alarm.sql
  #documentation = file("./cis_v130/docs/cis_v130_5_3.md")

  tags = merge(local.cis_v130_5_common_tags, {
    cis_item_id  = "5.3"
    cis_type     = "automated"
    cis_levels   = "1"
    cis_controls = "14.6"
  })
}

control "cis_v130_5_4" {
  title       = "5.4 Ensure routing tables for VPC peering are 'least access'"
  description = "A VPC comes with a default security group whose initial settings deny all inbound traffic, allow all outbound traffic, and allow all traffic between instances assigned to the security group. If you don't specify a security group when you launch an instance, the instance is automatically assigned to this default security group. Security groups provide stateful filtering of ingress/egress network traffic to AWS resources. It is recommended that the default security group restrict all traffic."
  sql         = query.alarm.sql
  #documentation = file("./cis_v130/docs/cis_v130_5_4.md")

  tags = merge(local.cis_v130_5_common_tags, {
    cis_item_id  = "5.4"
    cis_type     = "manual"
    cis_levels   = "1"
    cis_controls = "14.6"
  })
}

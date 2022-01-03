locals {
  pci_v321_es_common_tags = merge(local.pci_v321_common_tags, {
    service = "es"
  })
}

benchmark "pci_v321_es" {
  title         = "Elasticsearch"
  documentation = file("./pci_v321/docs/pci_v321_es.md")
  children = [
    control.pci_v321_es_1,
    control.pci_v321_es_2,
  ]
  tags = local.pci_v321_es_common_tags
}

control "pci_v321_es_1" {
  title         = "1 Amazon Elasticsearch Service domains should be in a VPC"
  description   = "This control checks whether Amazon Elasticsearch Service domains are in a VPC. It does not evaluate the VPC subnet routing configuration to determine public reachability."
  severity      = "critical"
  sql           = query.es_domain_in_vpc.sql
  documentation = file("./pci_v321/docs/pci_v321_es_1.md")

  tags = merge(local.pci_v321_es_common_tags, {
    pci_item_id      = "es_1"
    pci_requirements = "1.2.1,1.3.1,1.3.2,1.3.4,1.3.6"
  })
}

control "pci_v321_es_2" {
  title         = "2 Amazon Elasticsearch Service domains should have encryption at rest enabled"
  description   = "This control checks whether Amazon ES domains have encryption at rest configuration enabled."
  severity      = "medium"
  sql           = query.es_domain_encryption_at_rest_enabled.sql
  documentation = file("./pci_v321/docs/pci_v321_es_2.md")

  tags = merge(local.pci_v321_es_common_tags, {
    pci_item_id      = "es_2"
    pci_requirements = "3.4"
  })
}
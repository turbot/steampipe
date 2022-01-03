## Description

This control checks whether Amazon Elasticsearch Service domains are in a VPC.

It does not evaluate the VPC subnet routing configuration to determine public reachability.

This AWS control also does not check whether the Amazon ES resource-based policy permits public access by other accounts or external entities. You should ensure that Amazon ES domains are not attached to public subnets. See [Resource-based policies](https://docs.aws.amazon.com/elasticsearch-service/latest/developerguide/es-ac.html#es-ac-types-resource) in the Amazon Elasticsearch Service Developer Guide.

## Remediation

If you create a domain with a public endpoint, you cannot later place it within a VPC. Instead, you must create a new domain and migrate your data.

The reverse is also true. If you create a domain within a VPC, it cannot have a public endpoint. Instead, you must either create another domain or disable this control.

See the information on migrating from public access to VPC access in the [Amazon Elasticsearch Service Developer Guide](https://docs.aws.amazon.com/elasticsearch-service/latest/developerguide/es-vpc.html#es-migrating-public-to-vpc).
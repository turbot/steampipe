## Description

This control checks whether Amazon ES domains have encryption at rest configuration enabled.

## Remediation

By default, domains do not encrypt data at rest, and you cannot configure existing domains to use the feature.

To enable the feature, you must create another domain and migrate your data. For information about creating domains, see the [Amazon Elasticsearch Service Developer Guide](https://docs.aws.amazon.com/elasticsearch-service/latest/developerguide/es-createupdatedomains.html#es-createdomains).

Encryption of data at rest requires Amazon ES 5.1 or later. For more information about encrypting data at rest for Amazon ES, see the [Amazon Elasticsearch Service Developer Guide](https://docs.aws.amazon.com/elasticsearch-service/latest/developerguide/encryption-at-rest.html).
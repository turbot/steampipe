## Description

This control checks whether Amazon ES domains have node-to-node encryption enabled.

HTTPS (TLS) can be used to help prevent potential attackers from eavesdropping on or manipulating network traffic using person-in-the-middle or similar attacks. Only encrypted connections over HTTPS (TLS) should be allowed. Enabling node-to-node encryption for Amazon ES domains ensures that intra-cluster communications are encrypted in transit.

There can be a performance penalty associated with this configuration. You should be aware of and test the performance trade-off before enabling this option.

## Remediation

Node-to-node encryption can only be enabled on a new domain. To remediate this finding, first [Create a new domain](https://docs.aws.amazon.com/elasticsearch-service/latest/developerguide/es-createupdatedomains.html#es-createdomains) with the Node-to-node encryption check box selected. Then follow [Using a snapshot to migrate data](https://docs.aws.amazon.com/elasticsearch-service/latest/developerguide/es-version-migration.html#snapshot-based-migration) to migrate your data to the new domain.
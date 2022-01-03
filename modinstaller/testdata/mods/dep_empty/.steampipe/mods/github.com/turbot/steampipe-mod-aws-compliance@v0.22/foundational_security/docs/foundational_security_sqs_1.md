## Description

This control checks whether Amazon SQS queues are encrypted at rest.

Server-side encryption (SSE) allows you to transmit sensitive data in encrypted queues. To protect the content of messages in queues, SSE uses keys managed in AWS KMS. For more information, see [Encryption at rest](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-server-side-encryption.html) in the `Amazon Simple Queue Service Developer Guide`.

## Remediation

For information about managing SSE using the AWS Management Console, see [Configuring server-side encryption (SSE) for a queue (console)](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-configure-sse-existing-queue.html) in the Amazon Simple Queue Service Developer Guide.
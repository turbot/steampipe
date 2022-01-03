## Description

This control checks whether a Lambda function is configured with a dead-letter queue. The control fails if the Lambda function is not configured with a dead-letter queue.

As an alternative to an on-failure destination, you can configure your function with a dead-letter queue to save discarded events for further processing. A dead-letter queue acts the same as an on-failure destination. It is used when an event fails all processing attempts or expires without being processed.

A dead-letter queue allows you to look back at errors or failed requests to your Lambda function to debug or identify unusual behavior.

From a security perspective, it is important to understand why your function failed and to ensure that your function does not drop data or compromise data security as a result. For example, if your function cannot communicate to an underlying resource, that could be a symptom of a denial of service (DoS) attack elsewhere in the network.

## Remediation

You can configure a dead-letter queue from the AWS Lambda console.

**To configure a dead-letter queue**

1. Open the [AWS Lambda console](https://console.aws.amazon.com/lambda/.)

2. In the navigation pane, choose `Functions`.

3. Choose a function.

4. Choose `Configuration` and then choose `Asynchronous invocation`.

5. Under `Asynchronous invocation`, choose `Edit`.

6. Set `DLQ resource` to Amazon SQS or Amazon SNS.

7. Choose the target queue or topic.

8. Choose `Save`.
## Description

This control checks whether ACM certificates in your account are marked for
expiration within 30 days. It checks both imported certificates and
certificates provided by AWS Certificate Manager.

Certificates provided by ACM are automatically renewed. If you're using
certificates provided by ACM, you do not need to rotate SSL/TLS certificates.
ACM manages certificate renewals for you.

ACM does not automatically renew certificates that you import. You must renew
imported certificates manually.

## Remediation

ACM provides managed renewal for your Amazon issued SSL/TLS certificates. This
includes both public and private certificates issued by using ACM. If possible,
ACM renews your certificates automatically with no action required from you. A
certificate is eligible for renewal if it is associated with another AWS
service, such as Elastic Load Balancing or Amazon CloudFront. It can also be
renewed if it has been exported since being issued or last renewed.

If ACM cannot automatically validate one or more domain names in a certificate,
ACM notifies the domain owner that the domain must be validated manually. A
domain can require manual validation for the following reasons.

- ACM cannot establish an HTTPS connection with the domain.
- The certificate that is returned in the response to the HTTPS requests does not
match the one that ACM is renewing.

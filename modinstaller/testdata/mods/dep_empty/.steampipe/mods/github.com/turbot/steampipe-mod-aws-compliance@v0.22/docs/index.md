---
repository: "https://github.com/turbot/steampipe-mod-aws-compliance"
---

# AWS Compliance Mod

Run individual configuration, compliance and security controls or full compliance benchmarks for `Audit Manager Control Tower`, `AWS Foundational Security Best Practices`, `CIS`, `GDPR`, `HIPAA`, `NIST 800-53`, `NIST CSF`, `PCI DSS`, `RBI Cyber Security Framework` and `SOC 2` across all your AWS accounts.

![image](https://raw.githubusercontent.com/turbot/steampipe-mod-aws-compliance/main/docs/aws_cis_v140_console.png)

## References

[AWS](https://aws.amazon.com/) provides on-demand cloud computing platforms and APIs to authenticated customers on a metered pay-as-you-go basis.

[AWS Foundational Security Best Practices](https://docs.aws.amazon.com/securityhub/latest/userguide/securityhub-standards-fsbp-controls.html) is a set of controls that detect when your deployed accounts and resources deviate from security best practices.

[CIS AWS Benchmarks](https://www.cisecurity.org/benchmark/amazon_web_services/) provide a predefined set of compliance and security best-practice checks for AWS accounts.

[Audit Manager Control Tower](https://docs.aws.amazon.com/audit-manager/latest/userguide/controltower.html) provide the easiest way to set up and govern a secure, multi-account AWS environment.

[GDPR](https://docs.aws.amazon.com/audit-manager/latest/userguide/GDPR.html) provides a set of robust requirements that raise and harmonize standards for data protection, security, and compliance throughout the European Union (EU).

[HIPAA Compliance](https://aws.amazon.com/compliance/hipaa-compliance/) provides a set of general-purpose security standards for the U.S. Health Insurance Portability and Accountability Act (HIPAA).

[NIST 800-53](https://csrc.nist.gov/publications/detail/sp/800-53/rev-4/final) provides minimum baselines of security controls for U.S. federal information systems except those related to national security.

[NIST CSF](https://www.nist.gov/cyberframework) provides security standards for managing and reducing cybersecurity risk.

[PCI DSS](https://www.pcisecuritystandards.org) provides security standards for the payment card industry.

[RBI Cyber Security Framework](https://www.rbi.org.in/Scripts/NotificationUser.aspx?Id=11397) provides a cyber security framework for Urban Cooperative Banks (UCB) in India.

[SOC 2](https://docs.aws.amazon.com/audit-manager/latest/userguide/SOC2.html) provides an auditing procedure that ensures a company's data is securely managed.

[Steampipe](https://steampipe.io) is an open source CLI to instantly query cloud APIs using SQL.

[Steampipe Mods](https://steampipe.io/docs/reference/mod-resources#mod) are collections of `named queries`, and codified `controls` that can be used to test current configuration of your cloud resources against a desired configuration.

## Documentation

- **[Benchmarks and controls →](https://hub.steampipe.io/mods/turbot/aws_compliance/controls)**
- **[Named queries →](https://hub.steampipe.io/mods/turbot/aws_compliance/queries)**

## Get started

Install the AWS plugin with [Steampipe](https://steampipe.io):
```shell
steampipe plugin install aws
```

Clone:
```sh
git clone https://github.com/turbot/steampipe-mod-aws-compliance.git
cd steampipe-mod-aws-compliance
```

Run all benchmarks:
```shell
steampipe check all
```

Run a single benchmark:
```shell
steampipe check benchmark.cis_v140
```

Run a specific control:
```shell
steampipe check control.cis_v140_2_1_1
```

### Credentials

This mod uses the credentials configured in the [Steampipe AWS plugin](https://hub.steampipe.io/plugins/turbot/aws).

### Configuration

No extra configuration is required.

## Get involved

* Contribute: [GitHub Repo](https://github.com/turbot/steampipe-mod-aws-compliance)
* Community: [Slack Channel](https://join.slack.com/t/steampipe/shared_invite/zt-oij778tv-lYyRTWOTMQYBVAbtPSWs3g)

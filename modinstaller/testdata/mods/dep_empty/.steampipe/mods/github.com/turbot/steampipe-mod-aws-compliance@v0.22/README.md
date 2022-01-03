# AWS Compliance Scanning Tool

300+ checks covering industry defined security best practices across all AWS regions. Includes full support for multiple best practice benchmarks including PCI DSS, AWS Foundational Security, HIPAA, NIST 800-53, NIST CSF, Reserve Bank of India, Audit Manager Control Tower **and the latest (v1.4.0) CIS benchmarks**:

![image](https://raw.githubusercontent.com/turbot/steampipe-mod-aws-compliance/main/docs/aws_cis_v140_console.png)

Includes support for:
* [AWS CIS v1.3.0](https://hub.steampipe.io/mods/turbot/aws_compliance/controls/benchmark.cis_v130)
* [AWS CIS v1.4.0](https://hub.steampipe.io/mods/turbot/aws_compliance/controls/benchmark.cis_v140)
* [Audit Manager Control Tower](https://hub.steampipe.io/mods/turbot/aws_compliance/controls/benchmark.control_tower)
* [HIPAA](https://hub.steampipe.io/mods/turbot/aws_compliance/controls/benchmark.hipaa)
* [General Data Protection Regulation (GDPR)](https://hub.steampipe.io/mods/turbot/aws_compliance/controls/benchmark.gdpr) 🚀 New!
* [NIST 800-53 Revision 4](https://hub.steampipe.io/mods/turbot/aws_compliance/controls/benchmark.nist_800_53_rev_4)
* [NIST Cybersecurity Framework (CSF)](https://hub.steampipe.io/mods/turbot/aws_compliance/controls/benchmark.nist_csf)
* [PCI DSS v3.2.1](https://hub.steampipe.io/mods/turbot/aws_compliance/controls/benchmark.pci_v321)
* [AWS Foundational Security Best Practices](https://hub.steampipe.io/mods/turbot/aws_compliance/controls/benchmark.foundational_security)
* [Reserve Bank of India (RBI) Cyber Security Framework](https://hub.steampipe.io/mods/turbot/aws_compliance/controls/benchmark.rbi_cyber_security)
* [SOC 2](https://hub.steampipe.io/mods/turbot/aws_compliance/controls/benchmark.soc_2)

## Quick start

1) Download and install Steampipe (https://steampipe.io/downloads). Or use Brew:

```shell
brew tap turbot/tap
brew install steampipe

steampipe -v
steampipe version 0.9.1
```

2) Install the AWS plugin
```shell
steampipe plugin install aws
```

3) Clone this repo
```sh
git clone https://github.com/turbot/steampipe-mod-aws-compliance.git
cd steampipe-mod-aws-compliance
```

4) Generate your AWS credential report
```sh
aws iam generate-credential-report
```

5) Run all benchmarks:
```shell
steampipe check all
```

### Other things to checkout

Run an individual benchmark:
```shell
steampipe check benchmark.cis_v140
```

Use Steampipe introspection to view all current controls:
```
steampipe query "select resource_name from steampipe_control;"
```

Run a specific control:
```shell
steampipe check control.cis_v130_2_1_1
```

## Contributing

If you have an idea for additional compliance controls, or just want to help maintain and extend this mod ([or others](https://github.com/topics/steampipe-mod)) we would love you to join the community and start contributing. (Even if you just want to help with the docs.)

- **[Join our Slack community →](https://join.slack.com/t/steampipe/shared_invite/zt-oij778tv-lYyRTWOTMQYBVAbtPSWs3g)** and hang out with other Mod developers.
- **[Mod developer guide →](https://steampipe.io/docs/using-steampipe/writing-controls)**

Please see the [contribution guidelines](https://github.com/turbot/steampipe/blob/main/CONTRIBUTING.md) and our [code of conduct](https://github.com/turbot/steampipe/blob/main/CODE_OF_CONDUCT.md). All contributions are subject to the [Apache 2.0 open source license](https://github.com/turbot/steampipe-mod-aws-compliance/blob/main/LICENSE).

`help wanted` issues:
- [Steampipe](https://github.com/turbot/steampipe/labels/help%20wanted)
- [AWS Compliance Mod](https://github.com/turbot/steampipe-mod-aws-compliance/labels/help%20wanted)

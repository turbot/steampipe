import Report from "./index";
import { ReportContext } from "../../../../hooks/useReport";

const story = {
  title: "Layout/Report",
  component: Report,
};

export default story;

const Template = (args) => (
  <ReportContext.Provider value={{ dispatch: () => {}, report: args.report }}>
    <Report />
  </ReportContext.Provider>
);

export const Basic = Template.bind({});
Basic.args = {
  report: {
    name: "report.basic",
    title: "Basic Report",
    children: [
      {
        name: "text.header",
        type: "markdown",
        value: "## Basic Report",
      },
    ],
  },
};

export const TwoColumn = Template.bind({});
TwoColumn.args = {
  report: {
    name: "report.two_column",
    title: "Two Column Report",
    children: [
      {
        name: "text.header_1",
        type: "markdown",
        value: "## Column 1",
        width: 6,
      },
      {
        name: "text.header_2",
        type: "markdown",
        value: "## Column 2",
        width: 6,
      },
    ],
  },
};

export const LayoutContainer = Template.bind({});
LayoutContainer.args = {
  report: {
    name: "report.layout_container",
    title: "Layout Container Report",
    children: [
      {
        name: "container.wrapper",
        children: [
          {
            name: "text.title",
            type: "markdown",
            value: "## IAM Report",
          },
          {
            name: "chart.barchart",
            type: "bar",
            data: [
              ["Type", "Count"],
              ["User", 12],
              ["Policy", 93],
              ["Role", 48],
            ],
            title: "AWS IAM Entities",
          },
        ],
      },
    ],
  },
};

export const TwoColumnContainerLayout = Template.bind({});
TwoColumnContainerLayout.args = {
  report: {
    name: "report.layout_container",
    title: "Layout Container Report",
    children: [
      {
        name: "container.left",
        width: 6,
        children: [
          {
            name: "text.left_title",
            type: "markdown",
            value: "## Left",
          },
          {
            name: "chart.left_barchart",
            type: "bar",
            data: [
              ["Type", "Count"],
              ["User", 12],
              ["Policy", 93],
              ["Role", 48],
            ],
            title: "AWS IAM Entities",
          },
        ],
      },
      {
        name: "container.right",
        width: 6,
        children: [
          {
            name: "text.right_title",
            type: "markdown",
            value: "## Right",
          },
          {
            name: "chart.right_barchart",
            type: "line",
            data: [
              ["Type", "Count"],
              ["User", 12],
              ["Policy", 93],
              ["Role", 48],
            ],
            title: "AWS IAM Entities",
          },
        ],
      },
    ],
  },
};

// export const AwsCISSummary = Template.bind({});
// AwsCISSummary.storyName = "AWS CIS (Summary)";
// AwsCISSummary.args = {
//   report: {
//     title: "AWS CIS (Summary)",
//     children: [
//       {
//         name: "panel.title",
//         type: "markdown",
//         value: "# CIS Amazon Web Services Foundations Benchmark _v1.3.0 - 08-07-2020_",
//       },
//       {
//         id: "summary-stats",
//         panels: [
//           {
//             id: "compliant",
//             type: "counter",
//             query: "",
//             width: 3,
//             options: {
//               type: "ok",
//             },
//             data: [["Pass"], [51]],
//           },
//           {
//             id: "non-compliant",
//             type: "counter",
//             query: "",
//             width: 3,
//             options: {
//               type: "alert",
//             },
//             data: [["Fail"], [4]],
//           },
//           {
//             id: "summary-divider",
//             type: "markdown",
//             value: "---",
//           },
//         ],
//       },
//       {
//         id: "summary-control-table",
//         type: "control_table",
//         data: [
//           ["Status", "Section", "Check"],
//           [
//             "TBD",
//             "1 Identity and Access Management",
//             "1.1 Maintain current contact details (Manual)",
//           ],
//           [
//             "TBD",
//             "1 Identity and Access Management",
//             "1.2 Ensure security contact information is registered (Manual)",
//           ],
//           [
//             "TBD",
//             "1 Identity and Access Management",
//             "1.3 Ensure security questions are registered in the AWS account (Manual)",
//           ],
//           [
//             "ALARM",
//             "1 Identity and Access Management",
//             "1.4 Ensure no root user account access key exists (Automated)",
//           ],
//           [
//             "OK",
//             "1 Identity and Access Management",
//             '1.5 Ensure MFA is enabled for the "root user" account (Automated)',
//           ],
//           [
//             "OK",
//             "1 Identity and Access Management",
//             '1.6 Ensure hardware MFA is enabled for the "root user" account (Automated)',
//           ],
//           [
//             "ALARM",
//             "1 Identity and Access Management",
//             "1.7 Eliminate use of the root user for administrative and daily tasks (Automated)",
//           ],
//           [
//             "OK",
//             "1 Identity and Access Management",
//             "1.8 Ensure IAM password policy requires minimum length of 14 or greater (Automated)",
//           ],
//           [
//             "OK",
//             "1 Identity and Access Management",
//             "1.9 Ensure IAM password policy prevents password reuse (Automated)",
//           ],
//           [
//             "OK",
//             "1 Identity and Access Management",
//             "1.10 Ensure multi-factor authentication (MFA) is enabled for all IAM users that have a console password (Automated)",
//           ],
//           [
//             "TBD",
//             "1 Identity and Access Management",
//             "1.11 Do not setup access keys during initial user setup for all IAM users that have a console password (Manual)",
//           ],
//           [
//             "OK",
//             "1 Identity and Access Management",
//             "1.12 Ensure credentials unused for 90 days or greater are disabled (Automated)",
//           ],
//           [
//             "OK",
//             "1 Identity and Access Management",
//             "1.13 Ensure there is only one active access key available for any single IAM user (Automated)",
//           ],
//           [
//             "OK",
//             "1 Identity and Access Management",
//             "1.14 Ensure access keys are rotated every 90 days or less (Automated)",
//           ],
//           [
//             "OK",
//             "1 Identity and Access Management",
//             "1.15 Ensure IAM Users Receive Permissions Only Through Groups (Automated)",
//           ],
//           [
//             "OK",
//             "1 Identity and Access Management",
//             '1.16 Ensure IAM policies that allow full "*:*" administrative privileges are not attached (Automated)',
//           ],
//           [
//             "OK",
//             "1 Identity and Access Management",
//             "1.17 Ensure a support role has been created to manage incidents with AWS Support (Automated)",
//           ],
//           [
//             "TBD",
//             "1 Identity and Access Management",
//             "1.18 Ensure IAM instance roles are used for AWS resource access from instances (Manual)",
//           ],
//           [
//             "OK",
//             "1 Identity and Access Management",
//             "1.19 Ensure that all the expired SSL/TLS certificates stored in AWS IAM are removed (Automated)",
//           ],
//           [
//             "OK",
//             "1 Identity and Access Management",
//             "1.20 Ensure that S3 Buckets are configured with 'Block public access (bucket settings)' (Automated)",
//           ],
//           [
//             "OK",
//             "1 Identity and Access Management",
//             "1.21 Ensure that IAM Access analyzer is enabled (Automated)",
//           ],
//           [
//             "TBD",
//             "1 Identity and Access Management",
//             "1.22 Ensure IAM users are managed centrally via identity federation or AWS Organizations for multi-account environments (Manual)",
//           ],
//           [
//             "ALARM",
//             "2 Storage / 2.1 Simple Storage Service (S3)",
//             "2.1.1 Ensure all S3 buckets employ encryption-at-rest (Manual)",
//           ],
//           [
//             "OK",
//             "2 Storage / 2.1 Simple Storage Service (S3)",
//             "2.1.2 Ensure S3 Bucket Policy allows HTTPS requests (Manual)",
//           ],
//           [
//             "ALARM",
//             "2 Storage / 2.2 Elastic Compute Cloud (EC2)",
//             "2.1.2 Ensure EBS volume encryption is enabled (Manual)",
//           ],
//         ],
//       },
//     ],
//   },
// };
//
// export const AwsCIS2_1_1_Detail = Template.bind({});
// AwsCIS2_1_1_Detail.storyName = "AWS CIS 2.1.1 (Detail)";
// AwsCIS2_1_1_Detail.args = {
//   report: {
//     id: "aws-cis-2-1-1-detail",
//     panels: [
//       {
//         id: "header",
//         type: "markdown",
//         value: `# CIS Amazon Web Services Foundations Benchmark v1.3.0 - 08-07-2020
//
// ## 2 Storage
//
// ### 2.1 Simple Storage Service (S3)
//
// #### 2.1.1 Ensure all S3 buckets employ encryption-at-rest (Manual)`,
//       },
//       {
//         id: "overview",
//         type: "markdown",
//         value: `| Description | Rationale |
// |---|---|
// | Amazon S3 provides a variety of no, or low, cost encryption options to protect data at rest. | Encrypting data at rest reduces the likelihood that it is unintentionally exposed and can nullify the impact of disclosure if the encryption remains unbroken. |`,
//       },
//       {
//         id: "control-progress",
//         type: "control_progress",
//         data: [
//           ["Status", "Total"],
//           ["ok", 21],
//           ["ALARM", 14],
//         ],
//       },
//       {
//         id: "summary-divider",
//         type: "markdown",
//         value: "---",
//       },
//       {
//         id: "compliance-stats",
//         panels: [
//           {
//             id: "compliant",
//             type: "counter",
//             query: "",
//             width: 4,
//             options: {
//               type: "ok",
//             },
//             data: [["Compliant Buckets"], [21]],
//           },
//           {
//             id: "non-compliant",
//             type: "counter",
//             query: "",
//             width: 4,
//             options: {
//               type: "alert",
//             },
//             data: [["Non-Compliant Buckets"], [14]],
//           },
//           {
//             id: "summary-divider",
//             type: "markdown",
//             value: "---",
//           },
//         ],
//       },
//       {
//         id: "resources",
//         type: "markdown",
//         value: "### Bucket Compliance",
//         panels: [
//           {
//             id: "resources",
//             type: "control_table",
//             query: `select
//   CASE
//     when server_side_encryption_configuration is null then 'ALARM'
//     else 'OK'
//   END  as "Status",
//   name   as "Bucket",
//   region as "Region"
// from aws_s3_bucket
// order by
//   "Status",
//   "Bucket"`,
//             data: [
//               ["Status", "Bucket Name", "Region"],
//               [
//                 "ALARM",
//                 "aws-athena-query-results-876515858155-us-east-1",
//                 "us-east-1",
//               ],
//               [
//                 "ALARM",
//                 "aws-cloudtrail-logs-876515858155-8592de2c",
//                 "us-east-1",
//               ],
//               [
//                 "ALARM",
//                 "aws-cloudtrail-logs-876515858155-e0d83666",
//                 "us-east-1",
//               ],
//               ["ALARM", "deletesa11-test-bucket-venu-joe", "ca-central-1"],
//               [
//                 "ALARM",
//                 "elasticbeanstalk-eu-central-1-876515858155",
//                 "eu-central-1",
//               ],
//               ["ALARM", "my-demo-bucket-0011", "us-east-2"],
//               ["ALARM", "my-demo-bucket-111", "us-east-2"],
//               ["ALARM", "smyth-test-fluentd-logs", "us-east-1"],
//               ["ALARM", "vandelay-industries-cosmos-bucket", "us-east-1"],
//               ["ALARM", "vandelay-industries-darins-bucket", "us-east-1"],
//               ["ALARM", "vandelay-industries-elaines-bucket", "us-east-1"],
//               ["ALARM", "vandelay-industries-georges-bucket01", "us-east-1"],
//               ["ALARM", "vandelay-industries-vandelay01", "us-east-1"],
//               ["ALARM", "vanedaly-replicated-bucket-01", "us-east-1"],
//               ["OK", "deletesa11-streamalert-athena-results", "ca-central-1"],
//               ["OK", "deletesa11-streamalert-s3-logging", "ca-central-1"],
//               ["OK", "deletesa11-streamalert-terraform-state", "ca-central-1"],
//               ["OK", "deletesa11-streamalerts", "ca-central-1"],
//               ["OK", "jsmyth-test-bucket-8765", "us-east-1"],
//               ["OK", "turbot-876515858155-ap-northeast-1", "ap-northeast-1"],
//               ["OK", "turbot-876515858155-ap-northeast-2", "ap-northeast-2"],
//               ["OK", "turbot-876515858155-ap-south-1", "ap-south-1"],
//               ["OK", "turbot-876515858155-ap-southeast-1", "ap-southeast-1"],
//               ["OK", "turbot-876515858155-ap-southeast-2", "ap-southeast-2"],
//               ["OK", "turbot-876515858155-ca-central-1", "ca-central-1"],
//               ["OK", "turbot-876515858155-eu-central-1", "eu-central-1"],
//               ["OK", "turbot-876515858155-eu-north-1", "eu-north-1"],
//               ["OK", "turbot-876515858155-eu-west-1", "eu-west-1"],
//               ["OK", "turbot-876515858155-eu-west-2", "eu-west-2"],
//               ["OK", "turbot-876515858155-eu-west-3", "eu-west-3"],
//               ["OK", "turbot-876515858155-sa-east-1", "sa-east-1"],
//               ["OK", "turbot-876515858155-us-east-1", "us-east-1"],
//               ["OK", "turbot-876515858155-us-east-2", "us-east-2"],
//               ["OK", "turbot-876515858155-us-west-1", "us-west-1"],
//               ["OK", "turbot-876515858155-us-west-2", "us-west-2"],
//             ],
//           },
//         ],
//       },
//     ],
//   },
// };
//
// export const AwsCIS = Template.bind({});
// AwsCIS.storyName = "AWS CIS";
// AwsCIS.args = {
//   report: {
//     id: "aws-cis",
//     title: "AWS CIS",
//     panels: [
//       {
//         type: "markdown",
//         id: "header",
//         value: "# CIS Amazon Web Services Foundations Benchmark _v1.3.0 - 08-07-2020_",
//       },
//       {
//         id: "summary",
//         panels: [
//           {
//             id: "compliant",
//             type: "counter",
//             query: "",
//             width: 3,
//             options: {
//               type: "ok",
//             },
//             data: [["Compliant"], [52]],
//           },
//           {
//             id: "non-compliant",
//             type: "counter",
//             query: "",
//             width: 3,
//             options: {
//               type: "alert",
//             },
//             data: [["Non-Compliant"], [3]],
//           },
//           {
//             id: "summary-divider",
//             type: "markdown",
//             value: "---",
//           },
//         ],
//       },
//       {
//         id: "01",
//         panels: [
//           {
//             id: "01-summary",
//             panels: [
//               {
//                 id: "01-summary-title",
//                 type: "markdown",
//                 value: `## 1 Identity and Access Management
//
// This report contains recommendations for configuring identity and access management
// related options.`,
//                 width: 6,
//               },
//               {
//                 id: "01-summary-compliant",
//                 type: "counter",
//                 query: "",
//                 width: 3,
//                 options: {
//                   type: "ok",
//                 },
//                 data: [["Compliant"], [20]],
//               },
//               {
//                 id: "01-summary-non-compliant",
//                 type: "counter",
//                 query: "",
//                 width: 3,
//                 options: {
//                   type: "alert",
//                 },
//                 data: [["Non-Compliant"], [2]],
//               },
//             ],
//           },
//           // {
//           //   id: "01-01",
//           //   type: "markdown",
//           //   value: "## 1.1 Maintain current contact details (Manual)",
//           //   width: 4,
//           //   panels: [
//           //     {
//           //       id: "01-01-placeholder",
//           //       type: "placeholder",
//           //     },
//           //   ],
//           // },
//           // {
//           //   id: "01-02",
//           //   type: "markdown",
//           //   value: "## 1.2 Maintain current contact details (Manual)",
//           //   width: 4,
//           //   panels: [
//           //     {
//           //       id: "01-02-placeholder",
//           //       type: "placeholder",
//           //     },
//           //   ],
//           // },
//           // {
//           //   id: "01-03",
//           //   type: "markdown",
//           //   value:
//           //     "## 1.3 Ensure security questions are registered in the AWS account (Manual)",
//           //   width: 4,
//           //   panels: [
//           //     {
//           //       id: "01-03-placeholder",
//           //       type: "placeholder",
//           //     },
//           //   ],
//           // },
//           {
//             id: "01-04",
//             type: "markdown",
//             value: "## 1.4 Ensure no root user account access key exists (Automated) - Level 1",
//             panels: [
//               {
//                 id: "01-04-status",
//                 type: "status",
//                 query:
//                   "select account_access_keys_present from aws_iam_account_summary;",
//                 width: 1,
//                 data: [["alarm"]],
//               },
//               {
//                 id: "01-04-overview",
//                 type: "markdown",
//                 width: 11,
//                 value: `| Description | Rationale |
// |---|---|
// | The root user account is the most privileged user in an AWS account. AWS Access Keys provide programmatic access to a given AWS account. It is recommended that all access keys associated with the root user account be removed. | Removing access keys associated with the root user account limits vectors by which the account can be compromised. Additionally, removing the root access keys encourages the creation and use of role based accounts that are least privileged. |`,
//               },
//             ],
//           },
//           {
//             id: "01-05",
//             type: "markdown",
//             value: '## 1.5 Ensure MFA is enabled for the "root user" account (Automated) - Level 1',
//             panels: [
//               {
//                 id: "01-05-status",
//                 type: "status",
//                 query: `select
//   case
//     when account_mfa_enabled = 'true' then 'OK'
//     else 'ALARM'
//   end as "mfa_status"
// from
//   aws_iam_account_summary;`,
//                 width: 1,
//                 data: [["ok"]],
//               },
//               {
//                 id: "01-05-overview",
//                 type: "markdown",
//                 width: 11,
//                 value: `| Description | Rationale |
// |---|---|
// | The root user account is the most privileged user in an AWS account. Multi-factor Authentication (MFA) adds an extra layer of protection on top of a username and password. With MFA enabled, when a user signs in to an AWS website, they will be prompted for their username and password as well as for an authentication code from their AWS MFA device. **Note**: When virtual MFA is used for root accounts, it is recommended that the device used is NOT a personal device, but rather a dedicated mobile device (tablet or phone) that is managed to be kept charged and secured independent of any individual personal devices. ("non-personal virtual MFA") This lessens the risks of losing access to the MFA due to device loss, device trade-in or if the individual owning the device is no longer employed at the company. | Enabling MFA provides increased security for console access as it requires the authenticating principal to possess a device that emits a time-sensitive key and have knowledge of a credential. |`,
//               },
//             ],
//           },
//           {
//             id: "01-06",
//             type: "markdown",
//             value: '## 1.6 Ensure hardware MFA is enabled for the "root user" account (Automated) - Level 2',
//             panels: [
//               {
//                 id: "01-06-status",
//                 type: "status",
//                 query: `select
//   case
//     when account_mfa_enabled = 'true' then 'OK'
//     else 'ALARM'
//   end as "mfa_status"
// from
//   aws_iam_account_summary;`,
//                 width: 1,
//                 data: [["ok"]],
//               },
//               {
//                 id: "01-06-overview",
//                 type: "markdown",
//                 value: `| Description | Rationale |
// |---|---|
// | The root user account is the most privileged user in an AWS account. MFA adds an extra layer of protection on top of a user name and password. With MFA enabled, when a user signs in to an AWS website, they will be prompted for their user name and password as well as for an authentication code from their AWS MFA device. For Level 2, it is recommended that the root user account be protected with a hardware MFA. | A hardware MFA has a smaller attack surface than a virtual MFA. For example, a hardware MFA does not suffer the attack surface introduced by the mobile smartphone on which a virtual MFA resides. **Note**: Using hardware MFA for many, many AWS accounts may create a logistical device management issue. If this is the case, consider implementing this Level 2 recommendation selectively to the highest security AWS accounts and the Level 1 recommendation applied to the remaining accounts. |`,
//                 width: 11,
//               },
//             ],
//           },
//           {
//             id: "01-07",
//             type: "markdown",
//             value: "## 1.7 Eliminate use of the root user for administrative and daily tasks (Automated) - Level 1",
//             panels: [
//               {
//                 id: "01-07-status",
//                 type: "status",
//                 query: `select
//   case
//     when password_last_used is null
//     and access_key_1_last_used_date is null
//     and access_key_2_last_used_date is null then 'ok'
//     else 'alarm'
//   end as "status"
// from
//   aws_iam_credential_report
// where
//   user_name = '<root_account>';`,
//                 width: 1,
//                 data: [["status"], ["alarm"]],
//               },
//               {
//                 id: "01-07-overview",
//                 type: "markdown",
//                 value: `| Description | Rationale |
// |---|---|
// | With the creation of an AWS account, a _root user_ is created that cannot be disabled or deleted. That user has unrestricted access to and control over all resources in the AWS account. It is highly recommended that the use of this account be avoided for everyday tasks. | The root user has unrestricted access to and control over all account resources. Use of it is inconsistent with the principles of least privilege and separation of duties, and can lead to unnecessary harm due to error or account compromise. |`,
//                 width: 11,
//               },
//               {
//                 id: "01-07-controls",
//                 type: "control_table",
//                 data: [
//                   ["Status", "Check", "Last Used Date"],
//                   ["ALARM", "Root user password", "2021-03-01 16:00"],
//                   ["ALARM", "Root user access key 1", "2021-04-01 11:00"],
//                   ["OK", "Root user access key 2", "Never"],
//                 ],
//               },
//             ],
//           },
//           // {
//           //   id: "01-08",
//           //   type: "markdown",
//           //   value:
//           //     "## 1.8 Ensure IAM password policy requires minimum length of 14 or greater (Automated)",
//           //   panels: [
//           //     {
//           //       id: "01-08-data",
//           //       type: "control_list",
//           //       query:
//           //         "select\n  case\n    when minimum_password_length >= 14 then 'OK'\n    else 'ALARM'\n  end as status,\n  account_id as title,\n  'Minimum password length set to ' || minimum_password_length as reason\nfrom\n  aws_iam_account_password_policy\n",
//           //       width: 6,
//           //     },
//           //     {
//           //       id: "01-08-description",
//           //       type: "markdown",
//           //       value:
//           //         "#### Level 1\n\n#### Description:\n\nPassword policies are, in part, used to enforce password complexity requirements. IAM\npassword policies can be used to ensure password are at least a given length. It is\nrecommended that the password policy require a minimum password length 14.\n\n#### Rationale:\n\nSetting a password complexity policy increases account resiliency against brute force login\nattempts.\n",
//           //       width: 6,
//           //     },
//           //   ],
//           // },
//           // {
//           //   id: "01-10",
//           //   value:
//           //     "##\n## 1.10 Ensure multi-factor authentication (MFA) is enabled for all IAM users that have a console password\n",
//           //   type: "markdown",
//           //   panels: [
//           //     // {
//           //     //   id: "01-10-description",
//           //     //   type: "markdown",
//           //     //   options: {
//           //     //     display: "none",
//           //     //   },
//           //     //   value:
//           //     //     "| Level | Automated | Control | Description |\n| :-: | :-: | - | - |\n| 1 | Automated | 4.5 Use Multifactor Authentication For All Administrative Access | Multi-Factor Authentication (MFA) adds an extra layer of authentication assurance beyond traditional credentials. With MFA enabled, when a user signs in to the AWS Console, they will be prompted for their user name and password as well as for an authentication code from their physical or virtual MFA token. It is recommended that MFA be enabled for all accounts that have a console password. | \n",
//           //     //   width: 12,
//           //     // },
//           //     // {
//           //     //   id: "01-10-description2",
//           //     //   type: "markdown",
//           //     //   options: {
//           //     //     display: "none",
//           //     //   },
//           //     //   value:
//           //     //     "| Level | Automated | CIS Control |\n| :-: | :-: | - |\n| 1 | Automated | 4.5 Use Multifactor Authentication For All Administrative Access |\n",
//           //     //   width: 5,
//           //     // },
//           //     {
//           //       id: "01-10-details",
//           //       type: "markdown",
//           //       value:
//           //         "| Level | Automated |\n| :-: | :-: |\n| 1 | Automated |\n",
//           //       width: 4,
//           //     },
//           //     {
//           //       id: "01-10-detailed-content",
//           //       options: {
//           //         display: "block",
//           //       },
//           //       type: "markdown",
//           //       value:
//           //         "| CIS Control |\n| - |\n| 4.5 Use Multifactor Authentication For All Administrative Access |\n",
//           //       width: 8,
//           //     },
//           //     {
//           //       id: "01-10-description8",
//           //       options: {
//           //         display: "block",
//           //       },
//           //       width: 12,
//           //       panels: [
//           //         {
//           //           id: "01-10-description3",
//           //           type: "markdown",
//           //           value:
//           //             "| Description |\n| - |\n| Multi-Factor Authentication (MFA) adds an extra layer of authentication assurance beyond traditional credentials. With MFA enabled, when a user signs in to the AWS Console, they will be prompted for their user name and password as well as for an authentication code from their physical or virtual MFA token. It is recommended that MFA be enabled for all accounts that have a console password. | \n",
//           //           width: 6,
//           //         },
//           //         {
//           //           id: "01-10-description9",
//           //           type: "markdown",
//           //           value:
//           //             "| Rationale |\n| - |\n| Setting a password complexity policy increases account resiliency against brute force login attempts. |\n",
//           //           width: 6,
//           //         },
//           //       ],
//           //     },
//           //     // {
//           //     //   id: "01-10-detailed-content",
//           //     //   options: {
//           //     //     display: "none",
//           //     //   },
//           //     //   width: 8,
//           //     //   panels: [
//           //     //     {
//           //     //       type: "markdown",
//           //     //       id: "description3",
//           //     //       value:
//           //     //         "| Description |\n| - |\n| Multi-Factor Authentication (MFA) adds an extra layer of authentication assurance beyond traditional credentials. With MFA enabled, when a user signs in to the AWS Console, they will be prompted for their user name and password as well as for an authentication code from their physical or virtual MFA token. It is recommended that MFA be enabled for all accounts that have a console password. | \n",
//           //     //     },
//           //     //     {
//           //     //       type: "markdown",
//           //     //       id: "details2",
//           //     //       value:
//           //     //         "| CIS Control |\n| - |\n| 4.5 Use Multifactor Authentication For All Administrative Access |\n",
//           //     //     },
//           //     //   ],
//           //     // },
//           //     // {
//           //     //   id: "01-10-description5",
//           //     //   options: {
//           //     //     display: "none",
//           //     //   },
//           //     //   type: "markdown",
//           //     //   width: 8,
//           //     //   panels: [
//           //     //     {
//           //     //       value:
//           //     //         "| Description |\n| - |\n| Multi-Factor Authentication (MFA) adds an extra layer of authentication assurance beyond traditional credentials. With MFA enabled, when a user signs in to the AWS Console, they will be prompted for their user name and password as well as for an authentication code from their physical or virtual MFA token. It is recommended that MFA be enabled for all accounts that have a console password. | \n\n| CIS Control |\n| - |\n| 4.5 Use Multifactor Authentication For All Administrative Access |\n",
//           //     //     },
//           //     //   ],
//           //     // },
//           //     {
//           //       id: "01-10-data",
//           //       options: {
//           //         empty_value: "No results.",
//           //       },
//           //       type: "control_table",
//           //       query2:
//           //         "select\n  case\n    when not mfa_enabled then 'OK'\n    else 'ALARM'\n  end as \"Status\",\n  account_id as \"Account\",\n  name as \"User\",\n  case\n    when mfa_enabled then 'MFA is enabled'\n    else 'MFA not enabled'\n  end as \"Reason\"\nfrom\n  aws.aws_iam_user\n--where\n--  password_last_used is not null\n",
//           //       query:
//           //         'select\n  case\n    when path = \'/\' then \'OK\'\n    else \'ALARM\'\n  end as "Status",\n  account_id as "Account",\n  name as "Role",\n  description as "Reason"\nfrom\n  aws.aws_iam_role\norder by "Status"\n--where\n--  password_last_used is not null\n',
//           //       width: 12,
//           //     },
//           //     // {
//           //     //   id: "01-10-description",
//           //     //   type: "markdown",
//           //     //   options: {
//           //     //     display: "none",
//           //     //   },
//           //     //   value:
//           //     //     "| Level | Automated | Control | \n| :-: | :-: | - |\n| 1 | Automated | 4.5 Use Multifactor Authentication For All Administrative Access |\n\nMulti-Factor Authentication (MFA) adds an extra layer\nof authentication assurance beyond traditional credentials. With MFA\nenabled, when a user signs in to the AWS Console, they will be\nprompted for their user name and password as well as for an\nauthentication code from their physical or virtual MFA token. It is\nrecommended that MFA be enabled for all accounts that have a console\npassword.\n",
//           //     //   width: 6,
//           //     // },
//           //   ],
//           // },
//         ],
//       },
//     ],
//   },
// };
//
// export const AwsS3Top10 = Template.bind({});
// AwsS3Top10.storyName = "AWS S3 Top 10";
// AwsS3Top10.args = {
//   report: {
//     id: "aws-s3-top-10",
//     panels: [
//       {
//         id: "header",
//         type: "markdown",
//         value: "# AWS S3 Top 10",
//       },
//       {
//         id: "summary",
//         panels: [
//           {
//             id: "total-buckets",
//             width: 6,
//             type: "counter",
//             query: `select
//   count(*) as "Buckets"
// from
//   aws.aws_s3_bucket`,
//             data: [["Buckets"], [36]],
//           },
//           {
//             id: "public-buckets",
//             width: 6,
//             type: "counter",
//             query: `select
//   count(*) as "Public Buckets"
// from
//   aws.aws_s3_bucket
// where
//   bucket_policy_is_public`,
//             data: [["Public Buckets"], [4]],
//             options: {
//               type: "alert",
//             },
//           },
//         ],
//       },
//       {
//         id: "bucket-is-public",
//         type: "markdown",
//         value: "## 1. Public Buckets",
//         panels: [
//           {
//             id: "bucket-is-public-summary",
//             width: 5,
//             type: "markdown",
//             value: "List all buckets that are currently public.",
//           },
//           {
//             id: "bucket-is-public-data",
//             width: 7,
//             type: "table",
//             query: `select
//   name as "Public Buckets"
// from
//   aws_s3_bucket
// where
//   bucket_policy_is_public
// order by
//   name`,
//             data: [["Name"], ["foo"], ["bar"], ["bar-foo"], ["website"]],
//           },
//         ],
//       },
//       {
//         id: "bucket-encryption-at-rest",
//         type: "markdown",
//         value: "## 2. Encryption at Rest",
//         panels: [
//           {
//             id: "bucket-is-public-summary",
//             width: 5,
//             type: "markdown",
//             value: "List all buckets that are currently public.",
//           },
//           {
//             id: "bucket-is-public-data",
//             width: 7,
//             type: "table",
//             query: `select
//   name as "Public Buckets"
// from
//   aws_s3_bucket
// where
//   bucket_policy_is_public
// order by
//   name`,
//             data: [["Name"], ["foo"], ["bar"], ["bar-foo"], ["website"]],
//           },
//         ],
//       },
//     ],
//   },
// };
//
// export const KitchenSink = Template.bind({});
// KitchenSink.args = {
//   report: {
//     panels: [
//       {
//         id: "header",
//         type: "markdown",
//         value: "# AWS S3 Top 10",
//       },
//       {
//         id: "buckets-count",
//         type: "counter",
//         data: [["Buckets"], [55]],
//         options: {
//           type: "info",
//         },
//         width: 3,
//       },
//       {
//         id: "public-buckets-count",
//         type: "counter",
//         data: [["Public Buckets"], [7]],
//         options: {
//           type: "alert",
//         },
//         width: 3,
//       },
//       {
//         id: "description",
//         type: "markdown",
//         value: `**Description**
//
// Password policies are, in part, used to enforce password complexity requirements. IAM password policies can be used to ensure password are at least a given length. It is recommended that the password policy require a minimum password length 14.
//
// **Rationale**
//
// Setting a password complexity policy increases account resiliency against brute force login attempts.`,
//         width: 6,
//         height: 1,
//       },
//       {
//         id: "iam-entities-table",
//         type: "table",
//         data: [
//           ["Type", "Count"],
//           ["User", 12],
//           ["Policy", 93],
//           ["Role", 48],
//         ],
//         title: "AWS IAM Entities",
//       },
//       {
//         id: "bucket-control-table",
//         type: "control_table",
//         data: [
//           ["Status", "Account", "Bucket", "Reason"],
//           ["ALARM", "111122223333", "my-bucket", "Encryption not enabled."],
//           ["OK", "111122223333", "another-bucket", null],
//           ["ALARM", "444455556666", "test1", null],
//           ["ALARM", "111122223333", "long-bucket-name-for-testing", null],
//           ["ALARM", "111122223333", "more-buckets", null],
//           ["ALARM", "111122223333", "howyoudoingtoday", null],
//         ],
//       },
//       {
//         id: "s3-bucket-types-by-region",
//         type: "markdown",
//         value: "## AWS S3 Buckets by Region",
//         panels: [
//           {
//             id: "s3-bucket-types-by-region-barchart",
//             type: "barchart",
//             data: [
//               ["Region", "Public", "Non-Public"],
//               ["eu-west-1", 1, 26],
//               ["eu-west-2", 3, 8],
//               ["us-east-1", 2, 20],
//               ["us-east-2", 0, 10],
//               ["us-west-1", 1, 12],
//             ],
//           },
//         ],
//       },
//       {
//         id: "iam-entities",
//         type: "markdown",
//         value: "## AWS IAM Entities",
//         panels: [
//           {
//             id: "iam-entities-barchart",
//             type: "barchart",
//             data: [
//               ["Type", "Count"],
//               ["User", 12],
//               ["Policy", 93],
//               ["Role", 48],
//             ],
//           },
//         ],
//       },
//       {
//         id: "account-cost",
//         type: "markdown",
//         value: "## Account Cost per Day",
//         panels: [
//           {
//             id: "account-cost-linechart",
//             type: "linechart",
//             data: [
//               ["Date", "Cost ($)"],
//               ["2020-02-01", 84.45],
//               ["2020-02-02", 92.23],
//               ["2020-02-03", 101.3],
//               ["2020-02-04", 110.5],
//               ["2020-02-05", 174.95],
//               ["2020-02-06", 130.23],
//               ["2020-02-07", 150.2],
//               ["2020-02-08", 160],
//               ["2020-02-09", 172.12],
//               ["2020-02-10", 190],
//             ],
//           },
//         ],
//       },
//       {
//         id: "iam-entities-2",
//         type: "markdown",
//         value: "## AWS IAM Entities",
//         panels: [
//           {
//             id: "iam-entities-piechart",
//             type: "piechart",
//             data: [
//               ["Type", "Count"],
//               ["User", 12],
//               ["Policy", 93],
//               ["Role", 48],
//             ],
//             width: 4,
//           },
//         ],
//       },
//       {
//         id: "markdown-table",
//         type: "markdown",
//         width: 6,
//         value: `| Version | Published |
// | --- | --- |
// | v1.3.0 | 2020-07-08 |`,
//       },
//       {
//         id: "placeholder",
//         type: "placeholder",
//         width: 6,
//       },
//     ],
//   },
// };

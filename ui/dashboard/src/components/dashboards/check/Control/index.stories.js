import Text from "./index";
import { PanelStoryDecorator } from "../../../../utils/storybook";

const story = {
  title: "Primitives/Control",
  component: Text,
};

export default story;

const Template = (args) => (
  <PanelStoryDecorator definition={args} nodeType="control" />
);

export const Loading = Template.bind({});
Loading.args = {
  data: null,
  style: "info",
};

export const Error = Template.bind({});
Error.args = {
  data: null,
  error: "Something went wrong!",
};

export const Empty = Template.bind({});
Empty.args = {
  data: [],
};

export const SimpleDataFormat = Template.bind({});
SimpleDataFormat.args = {
  data: {
    group_id: "root_result_group",
    title: "",
    description: "",
    tags: {},
    summary: {
      status: {
        alarm: 0,
        ok: 1,
        info: 0,
        skip: 0,
        error: 0,
      },
    },
    groups: [],
    controls: [
      {
        control_id: "control.cis_v140_1_5",
        description:
          "The 'root' user account is the most privileged user in an AWS account. Multi-factor Authentication (MFA) adds an extra layer of protection on top of a username and password. With MFA enabled, when a user signs in to an AWS website, they will be prompted for their username and password as well as for an authentication code from their AWS MFA device.",
        severity: "",
        tags: {
          cis: "true",
          cis_item_id: "1.5",
          cis_level: "1",
          cis_section_id: "1",
          cis_type: "automated",
          cis_version: "v1.4.0",
          plugin: "aws",
          service: "iam",
        },
        title: "1.5 Ensure MFA is enabled for the 'root' user account",
        results: [
          {
            reason: "MFA enabled for root account.",
            resource: "arn:aws:::876515858155",
            status: "ok",
            dimensions: [
              {
                key: "account_id",
                value: "876515858155",
              },
            ],
          },
        ],
      },
    ],
  },
};
